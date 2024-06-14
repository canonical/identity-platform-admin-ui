// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type providerVerifierInterface interface {
	// Verify a raw string representing an authentication token
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type TokenVerifier interface {
	// VerifyAccessToken verifies a raw string representing a ory Hydra access token (either opaque or jwt)
	VerifyAccessToken(context.Context, string) (*Principal, error)
	// VerifyIDToken verifies a raw string representing a ory Hydra JWT ID token
	VerifyIDToken(context.Context, string) (*Principal, error)
}

type ProviderInterface interface {
	// Endpoint returns a set of endpoints from the well-known openid configuration
	Endpoint() oauth2.Endpoint
	// Verifier returns the token verifier associated with the specified OIDC issuer
	Verifier(*oidc.Config) *oidc.IDTokenVerifier
	// UserInfo returns user info object for the current authenticated user
	UserInfo(context.Context, oauth2.TokenSource) (*oidc.UserInfo, error)
}

type OAuth2ContextInterface interface {
	// Verifier returns an abstract TokenVerifier for Hydra token verification
	Verifier() TokenVerifier
	// LoginRedirect returns the URL preparaed for OAuth2 login process according to the authorization_code grant, and allows for specifying the nonce and state string values
	LoginRedirect(context.Context, string, string) string
	// RetrieveTokens performs the second leg of the OAuth2 authorization_code grant login flow
	RetrieveTokens(context.Context, string) (*oauth2.Token, error)
	// RefreshToken performs the OAuth2 refresh_token grant
	RefreshToken(context.Context, string) (*oauth2.Token, error)
}

type ReadableClaims interface {
	// Claims deserializes json fields in the struct passed as a parameter
	Claims(interface{}) error
}

type OAuth2HelperInterface interface {
	// RandomURLString generates a URL safe random string
	RandomURLString() string
}
type AuthCookieManagerInterface interface {
	SetNonceCookie(http.ResponseWriter, string, time.Duration)
	GetNonceCookie(*http.Request) string
	ClearNonceCookie(http.ResponseWriter)
	SetStateCookie(http.ResponseWriter, string, time.Duration)
	GetStateCookie(*http.Request) string
	ClearStateCookie(http.ResponseWriter)
}
