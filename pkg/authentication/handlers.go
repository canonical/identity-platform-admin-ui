// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

type Config struct {
	Enabled              bool          `validate:"required,boolean"`
	AuthCookieTTL        time.Duration `validate:"required,min=30s,max:1h"`
	issuer               string        `validate:"required"`
	clientID             string        `validate:"required"`
	clientSecret         string        `validate:"required"`
	redirectURL          string        `validate:"required"`
	verificationStrategy string        `validate:"required,oneof=jwks userinfo"`
	scopes               []string      `validate:"required,dive,required"`
}

type oauth2Tokens struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuthenticationConfig(enabled bool, issuer, clientID, clientSecret, redirectURL, verificationStrategy string, cookieTTL time.Duration, scopes []string) *Config {
	c := new(Config)
	c.Enabled = enabled

	c.issuer = issuer
	c.clientID = clientID
	c.clientSecret = clientSecret
	c.redirectURL = redirectURL
	c.verificationStrategy = verificationStrategy
	c.scopes = scopes
	c.AuthCookieTTL = cookieTTL

	return c
}

type API struct {
	apiKey           string
	payloadValidator validation.PayloadValidatorInterface
	oauth2           OAuth2ContextInterface
	helper           OAuth2HelperInterface
	authCookiesTTL   time.Duration

	tracer        trace.Tracer
	logger        logging.LoggerInterface
	cookieManager AuthCookieManagerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/login", a.handleLogin)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	// add the Otel HTTP Client
	r = r.WithContext(OtelHTTPClientContext(r.Context()))

	code := r.URL.Query().Get("code")
	if code == "" {
		// no code means login flow init
		a.oauth2.LoginRedirect(w, r)
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
		badRequest(w, errors.New("unable to retrieve ID token"))
		return
	}

	idToken, err := a.oauth2.Verifier().VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		a.logger.Errorf("unable to verify ID token, error: %v", err)
		badRequest(w, err)
		return
	}

	nonce := GetNonceCookie(r)
	if nonce == "" {
		a.logger.Error("nonce cookie not found")
		badRequest(w, errors.New("nonce cookie not found"))
		return
	}

	if idToken.Nonce != nonce {
		a.logger.Error("id token nonce does not match")
		badRequest(w, errors.New("id token nonce error"))
		return
	}

	ClearNonceCookie(w)

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

func NewAPI(oauth2Context OAuth2ContextInterface, helper OAuth2HelperInterface, cookieManager AuthCookieManagerInterface, authCookiesTTL time.Duration, tracer trace.Tracer, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "authentication"
	a.tracer = tracer
	a.logger = logger
	a.oauth2 = oauth2Context
	a.helper = helper
	a.authCookiesTTL = authCookiesTTL
	a.cookieManager = cookieManager

	return a
}
