// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

const (
	codeParameter  = "code"
	stateParameter = "state"
)

type Config struct {
	Enabled              bool     `validate:"required,boolean"`
	AuthCookieTTLSeconds int      `validate:"required"`
	issuer               string   `validate:"required"`
	clientID             string   `validate:"required"`
	clientSecret         string   `validate:"required"`
	redirectURL          string   `validate:"required"`
	verificationStrategy string   `validate:"required,oneof=jwks userinfo"`
	scopes               []string `validate:"required,dive,required"`
}

type oauth2Tokens struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuthenticationConfig(enabled bool, issuer, clientID, clientSecret, redirectURL, verificationStrategy string, cookieTTLSeconds int, scopes []string) *Config {
	c := new(Config)
	c.Enabled = enabled

	c.issuer = issuer
	c.clientID = clientID
	c.clientSecret = clientSecret
	c.redirectURL = redirectURL
	c.verificationStrategy = verificationStrategy
	c.scopes = scopes
	c.AuthCookieTTLSeconds = cookieTTLSeconds

	return c
}

type API struct {
	apiKey                string
	payloadValidator      validation.PayloadValidatorInterface
	oauth2                OAuth2ContextInterface
	helper                OAuth2HelperInterface
	cookieManager         AuthCookieManagerInterface
	authCookiesTTLSeconds int

	tracer trace.Tracer
	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/login", a.handleLogin)
	mux.Get("/api/v0/auth/callback", a.handleCallback)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	// add the Otel HTTP Client
	r = r.WithContext(OtelHTTPClientContext(r.Context()))

	nonce := a.helper.RandomURLString()
	state := a.helper.RandomURLString()

	ttl := time.Duration(a.authCookiesTTLSeconds) * time.Second

	a.cookieManager.SetNonceCookie(w, nonce, ttl)
	a.cookieManager.SetStateCookie(w, state, ttl)

	a.oauth2.LoginRedirect(w, r, nonce, state)
}

func (a *API) handleCallback(w http.ResponseWriter, r *http.Request) {
	// add the Otel HTTP Client
	r = r.WithContext(OtelHTTPClientContext(r.Context()))

	code := r.URL.Query().Get(codeParameter)
	if code == "" {
		a.logger.Error("OAuth2 code not found")
		badRequest(w, fmt.Errorf("OAuth2 code not found"))
		return
	}

	state := r.URL.Query().Get(stateParameter)
	if state == "" {
		a.logger.Error("OAuth2 state not found")
		badRequest(w, fmt.Errorf("OAuth2 state not found"))
		return
	}

	err := a.checkState(r, state)
	a.cookieManager.ClearStateCookie(w)
	if err != nil {
		badRequest(w, err)
		return
	}

	// else handle OAuth2 login second leg - retrieve tokens
	a.logger.Debugf("user login second leg with code '%s'", code)
	ctx := r.Context()

	oauth2Token, err := a.oauth2.RetrieveTokens(ctx, code)
	if err != nil {
		a.logger.Errorf("unable to retrieve tokens with code '%s', error: %v", code, err)
		badRequest(w, err)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		a.logger.Error("unable to retrieve ID token")
		badRequest(w, fmt.Errorf("unable to retrieve ID token"))
		return
	}

	idToken, err := a.oauth2.Verifier().VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		a.logger.Errorf("unable to verify ID token, error: %v", err)
		badRequest(w, err)
		return
	}

	err = a.checkNonce(r, idToken)
	a.cookieManager.ClearNonceCookie(w)
	if err != nil {
		badRequest(w, err)
		return
	}

	// TODO @barco: until we implement spec ID036 we just return tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	tokens := oauth2Tokens{
		IDToken:      rawIDToken,
		AccessToken:  oauth2Token.AccessToken,
		RefreshToken: oauth2Token.RefreshToken,
	}

	_ = json.NewEncoder(w).Encode(tokens)

}

func (a *API) checkNonce(r *http.Request, idToken *Principal) error {
	nonce := a.cookieManager.GetNonceCookie(r)
	if nonce == "" {
		a.logger.Error("nonce cookie not found")
		return fmt.Errorf("nonce cookie not found")
	}

	if idToken.Nonce != nonce {
		a.logger.Error("id token nonce does not match nonce cookie")
		return fmt.Errorf("id token nonce does not match nonce cookie")
	}

	return nil
}

func (a *API) checkState(r *http.Request, state string) error {
	stateCookieValue := a.cookieManager.GetStateCookie(r)
	if stateCookieValue == "" {
		a.logger.Error("state cookie not found")
		return fmt.Errorf("state cookie not found")
	}

	if stateCookieValue != state {
		a.logger.Error("state parameter does not match state cookie")
		return fmt.Errorf("state parameter does not match state cookie")
	}

	return nil
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(
		types.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		},
	)
	return
}

func NewAPI(authCookiesTTLSeconds int, oauth2Context OAuth2ContextInterface, helper OAuth2HelperInterface, cookieManager AuthCookieManagerInterface, tracer trace.Tracer, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "authentication"
	a.tracer = tracer
	a.logger = logger
	a.oauth2 = oauth2Context
	a.helper = helper
	a.authCookiesTTLSeconds = authCookiesTTLSeconds
	a.cookieManager = cookieManager

	return a
}
