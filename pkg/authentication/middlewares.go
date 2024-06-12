// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Middleware struct {
	allowListedEndpoints map[string]bool
	tracer               tracing.TracingInterface
	logger               logging.LoggerInterface
	oauth2               OAuth2ContextInterface
}

func (m *Middleware) SetAllowListedEndpoints(endpointsPrefixes ...string) {
	for _, prefix := range endpointsPrefixes {
		m.allowListedEndpoints[prefix] = true
	}
}

func (m *Middleware) isAllowListed(r *http.Request) bool {
	endpoint := r.URL.Path
	for prefix, _ := range m.allowListedEndpoints {
		if strings.HasPrefix(endpoint, prefix) {
			return true
		}
	}
	return false
}

func (m *Middleware) OAuth2Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "authentication.Middleware.OAuth2Authentication")
		defer span.End()
		r = r.WithContext(ctx)

		if m.isAllowListed(r) {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rawAccessToken, err := getBearerToken(r.Header)
		if err != nil {
			m.unauthorizedResponse(w, err)
			return
		}

		// add the Otel HTTP Client
		r.WithContext(OtelHTTPClientContext(r.Context()))

		principal, err := m.oauth2.Verifier().VerifyAccessToken(r.Context(), rawAccessToken)
		if err != nil {
			m.unauthorizedResponse(w, err)
			return
		}

		ctx = PrincipalContext(ctx, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) unauthorizedResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(types.Response{
		Status:  http.StatusUnauthorized,
		Message: fmt.Sprintf("unauthorized: %s", err.Error()),
	})
}

func getBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")

	if bearer == "" {
		return "", fmt.Errorf("bearer token is not present")
	}

	return strings.TrimPrefix(bearer, "Bearer "), nil
}

func NewAuthenticationMiddleware(oauth2 OAuth2ContextInterface, tracer tracing.TracingInterface, logger logging.LoggerInterface) *Middleware {
	m := new(Middleware)
	m.tracer = tracer
	m.logger = logger

	m.allowListedEndpoints = make(map[string]bool, 0)
	m.oauth2 = oauth2

	return m
}
