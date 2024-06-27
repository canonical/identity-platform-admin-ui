// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
)

const (
	codeParameter  = "code"
	stateParameter = "state"
)

type Config struct {
	Enabled                     bool     `validate:"required,boolean"`
	AuthCookieTTLSeconds        int      `validate:"required"`
	UserSessionCookieTTLSeconds int      `validate:"required"`
	CookiesEncryptionKey        string   `validate:"required,min=32,max=32"`
	issuer                      string   `validate:"required"`
	clientID                    string   `validate:"required"`
	clientSecret                string   `validate:"required"`
	redirectURL                 string   `validate:"required"`
	verificationStrategy        string   `validate:"required,oneof=jwks userinfo"`
	scopes                      []string `validate:"required,dive,required"`
}

func NewAuthenticationConfig(enabled bool, issuer, clientID, clientSecret, redirectURL, verificationStrategy string, authCookiesTTLSeconds, userSessionCookieTTLSeconds int, cookiesEncryptionKey string, scopes []string) *Config {
	c := new(Config)
	c.Enabled = enabled
	c.CookiesEncryptionKey = cookiesEncryptionKey

	c.issuer = issuer
	c.clientID = clientID
	c.clientSecret = clientSecret
	c.redirectURL = redirectURL
	c.verificationStrategy = verificationStrategy
	c.scopes = scopes
	c.AuthCookieTTLSeconds = authCookiesTTLSeconds
	c.UserSessionCookieTTLSeconds = userSessionCookieTTLSeconds

	return c
}

type API struct {
	apiKey           string
	payloadValidator validation.PayloadValidatorInterface
	oauth2           OAuth2ContextInterface
	helper           OAuth2HelperInterface
	cookieManager    AuthCookieManagerInterface

	tracer trace.Tracer
	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/auth", a.handleLogin)
	mux.Get("/api/v0/auth/callback", a.handleCallback)
	mux.Get("/api/v0/auth/me", a.handleMe)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	// add the Otel HTTP Client
	r = r.WithContext(OtelHTTPClientContext(r.Context()))

	nonce := a.helper.RandomURLString()
	state := a.helper.RandomURLString()

	a.cookieManager.SetNonceCookie(w, nonce)
	a.cookieManager.SetStateCookie(w, state)

	redirect := a.oauth2.LoginRedirect(r.Context(), nonce, state)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (a *API) handleCallback(w http.ResponseWriter, r *http.Request) {
	// add the Otel HTTP Client
	r = r.WithContext(OtelHTTPClientContext(r.Context()))

	code := r.URL.Query().Get(codeParameter)
	if code == "" {
		a.logger.Error("OAuth2 code not found")
		a.badRequest(w, fmt.Errorf("OAuth2 code not found"))
		return
	}

	state := r.URL.Query().Get(stateParameter)
	if state == "" {
		a.logger.Error("OAuth2 state not found")
		a.badRequest(w, fmt.Errorf("OAuth2 state not found"))
		return
	}

	err := a.checkState(r, state)
	a.cookieManager.ClearStateCookie(w)
	if err != nil {
		a.badRequest(w, err)
		return
	}

	// else handle OAuth2 login second leg - retrieve tokens
	a.logger.Debugf("user login second leg with code '%s'", code)
	ctx := r.Context()

	oauth2Token, err := a.oauth2.RetrieveTokens(ctx, code)
	if err != nil {
		a.logger.Errorf("unable to retrieve tokens with code '%s', error: %v", code, err)
		a.badRequest(w, err)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		a.logger.Error("unable to retrieve ID token")
		a.badRequest(w, fmt.Errorf("unable to retrieve ID token"))
		return
	}

	idToken, err := a.oauth2.Verifier().VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		a.logger.Errorf("unable to verify ID token, error: %v", err)
		a.badRequest(w, err)
		return
	}

	err = a.checkNonce(r, idToken)
	a.cookieManager.ClearNonceCookie(w)
	if err != nil {
		a.badRequest(w, err)
		return
	}

	a.cookieManager.SetIDTokenCookie(w, rawIDToken)
	a.cookieManager.SetAccessTokenCookie(w, oauth2Token.AccessToken)
	a.cookieManager.SetRefreshTokenCookie(w, oauth2Token.RefreshToken)

	http.Redirect(w, r, ui.UIPrefix, http.StatusFound)
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

func (a *API) badRequest(w http.ResponseWriter, err error) {
	a.cookieManager.ClearNonceCookie(w)
	a.cookieManager.ClearStateCookie(w)

	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(
		types.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		},
	)
	return
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	// if we got here then principal must be populated by the middleware chain
	principal := PrincipalFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(principal)
	if err != nil {
		// this should never happen
		a.logger.Errorf("error serializing user me response, %v", err)
		a.internalServerError(w, err)
		return
	}
}

func (a *API) internalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(
		types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		},
	)
}

func NewAPI(
	oauth2Context OAuth2ContextInterface,
	helper OAuth2HelperInterface,
	cookieManager AuthCookieManagerInterface,
	tracer trace.Tracer,
	logger logging.LoggerInterface,
) *API {
	a := new(API)
	a.apiKey = "authentication"
	a.oauth2 = oauth2Context
	a.helper = helper
	a.cookieManager = cookieManager

	a.logger = logger
	a.tracer = tracer

	return a
}
