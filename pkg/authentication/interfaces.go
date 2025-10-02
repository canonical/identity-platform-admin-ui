// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	hClient "github.com/ory/hydra-client-go/v2"
	"golang.org/x/oauth2"
)

type providerVerifierInterface interface {
	// Verify a raw string representing an authentication token
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type TokenVerifier interface {
	// VerifyAccessToken verifies a raw string representing a ory Hydra access token (either opaque or jwt)
	VerifyAccessToken(context.Context, string) (*ServicePrincipal, error)
	// VerifyIDToken verifies a raw string representing a ory Hydra JWT ID token
	VerifyIDToken(context.Context, string) (*UserPrincipal, error)
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
	// Logout performs session and tokens revocation against the Hydra Admin APIs
	Logout(ctx context.Context, principal PrincipalInterface) error
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
	// SetNonceCookie sets the nonce cookie on the response with the specified duration as MaxAge
	SetNonceCookie(http.ResponseWriter, string)
	// GetNonceCookie returns the string value of the nonce cookie if present, or empty string otherwise
	GetNonceCookie(*http.Request) string
	// ClearNonceCookie sets the expiration of the cookie to epoch
	ClearNonceCookie(http.ResponseWriter)
	// SetStateCookie  sets the nonce cookie on the response with the specified duration as MaxAge
	SetStateCookie(http.ResponseWriter, string)
	// GetStateCookie returns the string value of the state cookie if present, or empty string otherwise
	GetStateCookie(*http.Request) string
	// ClearStateCookie sets the expiration of the cookie to epoch
	ClearStateCookie(http.ResponseWriter)
	// SetIDTokenCookie sets the encrypted ID token value cookie
	SetIDTokenCookie(http.ResponseWriter, string)
	// GetIDTokenCookie returns the string value of the ID token cookie if present, or empty string otherwise
	GetIDTokenCookie(*http.Request) string
	// ClearIDTokenCookie sets the expiration of the cookie to epoch
	ClearIDTokenCookie(http.ResponseWriter)
	// SetAccessTokenCookie sets the encrypted access token value cookie
	SetAccessTokenCookie(http.ResponseWriter, string)
	// GetAccessTokenCookie returns the string value of the access token cookie if present, or empty string otherwise
	GetAccessTokenCookie(*http.Request) string
	// ClearAccessTokenCookie sets the expiration of the cookie to epoch
	ClearAccessTokenCookie(http.ResponseWriter)
	// SetRefreshTokenCookie sets the encrypted refresh token value cookie
	SetRefreshTokenCookie(http.ResponseWriter, string)
	// GetRefreshTokenCookie returns the string value of the refresh token cookie if present, or empty string otherwise
	GetRefreshTokenCookie(*http.Request) string
	// ClearRefreshTokenCookie sets the expiration of the cookie to epoch
	ClearRefreshTokenCookie(http.ResponseWriter)
	// SetNextToCookie sets the encrypted nextTo relative url value cookie
	SetNextToCookie(http.ResponseWriter, string)
	// GetNextToCookie  returns the string value of the nextTo cookie if present, or empty string otherwise
	GetNextToCookie(*http.Request) string
	// ClearNextToCookie sets the expiration of the cookie to epoch
	ClearNextToCookie(http.ResponseWriter)
}

type EncryptInterface interface {
	// Encrypt a plain text string, returns the encrypted string in hex format or an error
	Encrypt(string) (string, error)
	// Decrypt a hex string, returns the decrypted string or an error
	Decrypt(string) (string, error)
}

type HTTPClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type PrincipalInterface interface {
	Identifier() string
	Session() string
	AccessToken() string
	RefreshToken() string
	IDToken() string
}

type SessionManagerInterface interface {
	GetIdentitySession(context.Context, []*http.Cookie) (*SessionData, error)
	DisableSession(ctx context.Context, sessionID string) (*SessionData, error)
}

type HydraClientInterface interface {
	OAuth2Api() hClient.OAuth2Api
}
