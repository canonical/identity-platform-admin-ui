// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Middleware struct {
	allowListedEndpoints map[string]bool
	oauth2               OAuth2ContextInterface
	cookieManager        AuthCookieManagerInterface

	tracer tracing.TracingInterface
	logger logging.LoggerInterface
}

type cookieTokens struct {
	accessToken  string
	idToken      string
	refreshToken string
}

func (m *Middleware) SetAllowListedEndpoints(endpointsPrefixes ...string) {
	for _, prefix := range endpointsPrefixes {
		m.allowListedEndpoints[prefix] = true
	}
}

func (m *Middleware) isAllowListed(r *http.Request) bool {
	endpoint := r.URL.Path
	_, ok := m.allowListedEndpoints[endpoint]
	return ok
}

func (m *Middleware) OAuth2AuthenticationChain() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		m.oAuth2BearerAuthentication,
		m.oAuth2CookieAuthentication,
	}
}

func (m *Middleware) oAuth2BearerAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "authentication.Middleware.oAuth2BearerAuthentication")
		defer span.End()

		if m.isAllowListed(r) {
			next.ServeHTTP(w, r)
			return
		}

		var (
			servicePrincipal *ServicePrincipal
			err              error
		)

		// add the Otel HTTP Client
		r = r.WithContext(OtelHTTPClientContext(r.Context()))

		if rawAccessToken, found := m.getBearerToken(r.Header); found {
			servicePrincipal, err = m.oauth2.Verifier().VerifyAccessToken(r.Context(), rawAccessToken)
			if err != nil {
				m.unauthorizedResponse(w, err)
				return
			}

			servicePrincipal.RawAccessToken = rawAccessToken
		}

		if servicePrincipal != nil {
			ctx = PrincipalContext(ctx, servicePrincipal)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) oAuth2CookieAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "authentication.Middleware.oAuth2CookieAuthentication")
		defer span.End()

		if m.isAllowListed(r) {
			next.ServeHTTP(w, r)
			return
		}

		err := fmt.Errorf("no authentication token found")

		// request context also carries over the OTEL http client set previously in the chain
		if PrincipalFromContext(r.Context()) != nil {
			// principal != nil means bearer authentication set the principal and we can move on
			next.ServeHTTP(w, r)
			return
		}

		var userPrincipal *UserPrincipal

		if tokens, foundAny := m.getCookieTokens(r); foundAny &&
			tokens.accessToken != "" && tokens.idToken != "" {
			// we first validate accessToken, but rely on ID token validity and info to populate UserPrincipal attributes
			_, errAccessToken := m.oauth2.Verifier().VerifyAccessToken(r.Context(), tokens.accessToken)
			userPrincipal, err = m.oauth2.Verifier().VerifyIDToken(r.Context(), tokens.idToken)
			if userPrincipal != nil {
				userPrincipal.RawIdToken = tokens.idToken
				userPrincipal.RawAccessToken = tokens.accessToken
				userPrincipal.RawRefreshToken = tokens.refreshToken
			}

			// we give precedence to access token errors by overwriting ID token error if any
			if errAccessToken != nil {
				err = errAccessToken
			}

			if err != nil && m.shouldRefresh(err, tokens) {
				// if the error is a TokenExpiredError and the refresh token is available, we try to refresh it
				userPrincipal, err = m.performRefresh(w, r, tokens)
			}
		}

		if err != nil {
			m.unauthorizedResponse(w, fmt.Errorf("unable to authenticate from either bearer or cookie token, %v", err))
			return
		}

		ctx = PrincipalContext(ctx, userPrincipal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) clearTokensCookies(w http.ResponseWriter) {
	m.cookieManager.ClearIDTokenCookie(w)
	m.cookieManager.ClearAccessTokenCookie(w)
	m.cookieManager.ClearRefreshTokenCookie(w)
}

func (m *Middleware) shouldRefresh(err error, tokens *cookieTokens) bool {
	var expiredError *oidc.TokenExpiredError
	return errors.As(err, &expiredError) && tokens.refreshToken != ""
}

func (m *Middleware) performRefresh(w http.ResponseWriter, r *http.Request, cookieTokens *cookieTokens) (*UserPrincipal, error) {
	tokens, err := m.oauth2.RefreshToken(r.Context(), cookieTokens.refreshToken)
	if err != nil {
		return nil, err
	}

	rawIdToken := tokens.Extra("id_token").(string)

	// overwrite existing tokens cookies
	m.cookieManager.SetIDTokenCookie(w, rawIdToken)
	m.cookieManager.SetAccessTokenCookie(w, tokens.AccessToken)
	m.cookieManager.SetRefreshTokenCookie(w, tokens.RefreshToken)

	// get a populated UserPrincipal object from the ID Token
	userPrincipal, err := m.oauth2.Verifier().VerifyIDToken(r.Context(), rawIdToken)
	if err != nil {
		return nil, err
	}

	userPrincipal.RawIdToken = rawIdToken
	userPrincipal.RawAccessToken = tokens.AccessToken
	userPrincipal.RawRefreshToken = tokens.RefreshToken

	return userPrincipal, nil
}

func (m *Middleware) getCookieTokens(r *http.Request) (*cookieTokens, bool) {
	ret := new(cookieTokens)
	foundAny := false

	idToken := m.cookieManager.GetIDTokenCookie(r)
	if idToken != "" {
		foundAny = true
		ret.idToken = idToken
	}

	accessToken := m.cookieManager.GetAccessTokenCookie(r)
	if accessToken != "" {
		foundAny = true
		ret.accessToken = accessToken
	}

	refreshToken := m.cookieManager.GetRefreshTokenCookie(r)
	if refreshToken != "" {
		foundAny = true
		ret.refreshToken = refreshToken
	}

	return ret, foundAny
}

func (m *Middleware) getBearerToken(headers http.Header) (string, bool) {
	bearer := headers.Get("Authorization")
	if bearer == "" {
		return "", false
	}

	return strings.TrimPrefix(bearer, "Bearer "), true
}

func (m *Middleware) unauthorizedResponse(w http.ResponseWriter, err error) {
	// in case of any unauthorized response we clear all cookies
	m.clearTokensCookies(w)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(types.Response{
		Status:  http.StatusUnauthorized,
		Message: fmt.Sprintf("unauthorized: %s", err.Error()),
	})
}

func NewAuthenticationMiddleware(oauth2 OAuth2ContextInterface, cookieManager AuthCookieManagerInterface, tracer tracing.TracingInterface, logger logging.LoggerInterface) *Middleware {
	m := new(Middleware)

	m.allowListedEndpoints = make(map[string]bool, 0)
	m.oauth2 = oauth2
	m.cookieManager = cookieManager

	m.tracer = tracer
	m.logger = logger
	return m
}

func AuthenticationDisabledMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		noAuthPrincipal := UserPrincipal{
			Subject: "0000000000000000",
			Name:    "User",
			Email:   "user@email.com",
		}

		next.ServeHTTP(w, r.WithContext(PrincipalContext(r.Context(), &noAuthPrincipal)))
	})
}
