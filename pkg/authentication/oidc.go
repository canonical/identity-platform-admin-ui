// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

type principalContextKey int

var (
	PrincipalContextKey principalContextKey
	otelHTTPClient      = http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
)

type Principal struct {
	Subject string `json:"sub"`
	Nonce   string `json:"nonce"`
}

func NewPrincipalFromClaims(c ReadableClaims) (*Principal, error) {
	a := new(Principal)
	if err := c.Claims(a); err != nil {
		return nil, err
	}
	return a, nil
}

func PrincipalContext(ctx context.Context, principal *Principal) context.Context {
	parent := ctx
	if ctx == nil {
		parent = context.Background()
	}

	return context.WithValue(parent, PrincipalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) *Principal {
	if ctx == nil {
		return nil
	}

	return ctx.Value(PrincipalContextKey).(*Principal)
}

func OtelHTTPClientContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, otelHTTPClient)
}

type OIDCProviderSupplier = func(ctx context.Context, issuer string) (*oidc.Provider, error)

type OAuth2Context struct {
	client   *oauth2.Config
	verifier TokenVerifier

	tracer  trace.Tracer
	logger  logging.LoggerInterface
	monitor monitoring.MonitorInterface
}

func (o *OAuth2Context) LoginRedirect(ctx context.Context, nonce, state string) string {
	_, span := o.tracer.Start(ctx, "authentication.OAuth2Context.LoginRedirect")
	defer span.End()

	// TODO: remove `audience` parameter when https://github.com/canonical/identity-platform-login-ui/issues/244 is addressed
	return o.client.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.SetAuthURLParam("audience", o.client.ClientID))
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

	return o
}
