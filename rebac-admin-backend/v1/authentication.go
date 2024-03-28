// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// getAuthenticationMiddleware returns a middleware function that delegates the
// extraction of the caller identity to the provided authenticator backend, and
// store the returned identity in the request context.
// If no authenticator backend is provided, a no-op middleware is returned.
func (b *ReBACAdminBackend) authenticationMiddleware() resources.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b.params.Authenticator == nil {
				// This should never happen, because the outmost constructor does not
				// allow nil for authenticator. But it's possible to miss this requirement
				// in manually created instances (like in tests), we should do the checking.
				writeErrorResponse(w, NewUnknownError("missing authenticator"))
				return
			}

			identity, err := b.params.Authenticator.Authenticate(r)
			if err != nil {
				writeServiceErrorResponse(w, b.params.AuthenticatorErrorMapper, err)
				return
			}
			if identity == nil {
				writeErrorResponse(w, NewAuthenticationError("nil identity"))
				return
			}
			next.ServeHTTP(w, newRequestWithIdentityInContext(r, identity))
		})
	}
}

type authenticatedIdentityContextKey struct{}

// GetIdentityFromContext fetches authenticated identity of the caller from the
// given request context. If the value was not found in the given context, this
// will return an error.
//
// The function is intended to be used by service backends.
func GetIdentityFromContext(ctx context.Context) (any, error) {
	identity := ctx.Value(authenticatedIdentityContextKey{})
	if identity == nil {
		return nil, NewAuthenticationError("missing caller identity")
	}
	return identity, nil
}

// newRequestWithIdentityInContext sets the given authenticated identity in a
// new request instance context and returns the new request.
func newRequestWithIdentityInContext(r *http.Request, identity any) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authenticatedIdentityContextKey{}, identity))
}
