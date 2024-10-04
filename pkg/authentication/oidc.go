// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	client "github.com/ory/hydra-client-go/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
)

var (
	otelHTTPClient = http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
)

func OtelHTTPClientContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, otelHTTPClient)
}

type OIDCProviderSupplier = func(ctx context.Context, issuer string) (*oidc.Provider, error)

type OAuth2Context struct {
	client      *oauth2.Config
	verifier    TokenVerifier
	hydraAdmin  clients.HydraClientInterface
	hydraPublic clients.HydraClientInterface

	tracer  trace.Tracer
	logger  logging.LoggerInterface
	monitor monitoring.MonitorInterface
}

type hydraAPIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (o *OAuth2Context) LoginRedirect(ctx context.Context, nonce, state string) string {
	_, span := o.tracer.Start(ctx, "authentication.OAuth2Context.LoginRedirect")
	defer span.End()

	return o.client.AuthCodeURL(state, oidc.Nonce(nonce))
}

func (o *OAuth2Context) RetrieveTokens(ctx context.Context, code string) (*oauth2.Token, error) {
	_, span := o.tracer.Start(ctx, "authentication.OAuth2Context.RetrieveTokens")
	defer span.End()

	return o.client.Exchange(ctx, code)
}

func (o *OAuth2Context) RefreshToken(ctx context.Context, rawRefreshToken string) (*oauth2.Token, error) {
	_, span := o.tracer.Start(ctx, "authentication.OAuth2Context.RefreshToken")
	defer span.End()

	return o.client.
		TokenSource(ctx, &oauth2.Token{RefreshToken: rawRefreshToken}).
		Token()
}

func (o *OAuth2Context) Logout(ctx context.Context, principal PrincipalInterface) error {
	_, span := o.tracer.Start(ctx, "authentication.OAuth2Context.Logout")
	defer span.End()

	if principal == nil {
		return fmt.Errorf("no principal provided")
	}

	err := o.revokeSession(ctx, principal)
	if err != nil {
		return err
	}

	return o.revokeToken(ctx, principal)
}

func (o *OAuth2Context) revokeSession(ctx context.Context, principal PrincipalInterface) error {
	// in case of a CLI user no SessionID is present
	if principal.Session() == "" {
		return nil
	}

	req := o.hydraAdmin.OAuth2Api().
		RevokeOAuth2LoginSessions(ctx).
		Sid(principal.Session())

	response, err := o.hydraAdmin.
		OAuth2Api().
		RevokeOAuth2LoginSessionsExecute(req)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return o.hydraApiError("revoke session", response)
	}

	return nil
}

func (o *OAuth2Context) revokeToken(ctx context.Context, principal PrincipalInterface) error {
	token := principal.AccessToken()
	if principal.RefreshToken() != "" {
		token = principal.RefreshToken()
	}

	ctx = context.WithValue(ctx, client.ContextBasicAuth, client.BasicAuth{
		UserName: o.client.ClientID,
		Password: o.client.ClientSecret,
	})

	req := o.hydraPublic.OAuth2Api().
		RevokeOAuth2Token(ctx).
		Token(token)

	response, err := o.hydraPublic.
		OAuth2Api().
		RevokeOAuth2TokenExecute(req)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return o.hydraApiError("revoke token", response)
	}

	return nil
}

func (o *OAuth2Context) hydraApiError(requestName string, resp *http.Response) error {
	revokeErr := new(hydraAPIError)
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(revokeErr); err != nil {
		return err
	}

	return fmt.Errorf("%s request failed, error: %s, description: %s", requestName, revokeErr.Error, revokeErr.ErrorDescription)
}

func (o *OAuth2Context) Verifier() TokenVerifier {
	return o.verifier
}

func NewOAuth2Context(config *Config, getProvider OIDCProviderSupplier, tracer trace.Tracer, logger logging.LoggerInterface, monitor monitoring.MonitorInterface) *OAuth2Context {
	o := new(OAuth2Context)
	o.tracer = tracer
	o.logger = logger
	o.monitor = monitor

	ctx := OtelHTTPClientContext(context.Background())

	provider, err := getProvider(ctx, config.issuer)
	if err != nil {
		o.logger.Fatalf("Unable to fetch provider info, error: %v", err.Error())
	}

	var verifier TokenVerifier
	switch config.verificationStrategy {
	case "jwks":
		verifier = NewJWKSTokenVerifier(provider, config.clientID, tracer, logger, monitor)
	case "userinfo":
		verifier = NewUserinfoTokenVerifier(provider, config.clientID, tracer, logger, monitor)
	default:
		o.logger.Fatalf("OAuth2VerificationStrategy value is not valid, expected one of 'jwks, userinfo', got %v", config.verificationStrategy)
	}

	o.verifier = verifier
	o.client = &oauth2.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		RedirectURL:  config.redirectURL,

		Endpoint: provider.Endpoint(),
		Scopes:   config.scopes,
	}

	o.hydraAdmin = config.hydraAdminAPIClient
	o.hydraPublic = config.hydraPublicAPIClient

	return o
}
