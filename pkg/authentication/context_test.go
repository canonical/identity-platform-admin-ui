// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go

func TestNewPrincipalFromClaims(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClaims := NewMockReadableClaims(ctrl)

	mockClaims.EXPECT().Claims(gomock.Any()).Times(1).
		DoAndReturn(
			func(p *Principal) error {
				p.Subject = "mock-sub"
				return nil
			})

	principal, err := NewPrincipalFromClaims(mockClaims)

	if err != nil {
		t.Fatalf("returned error should be null, but it is not")
	}

	if principal.Subject != "mock-sub" {
		t.Fatalf("returned principal subject doesn't match expected")
	}
}

func TestNewPrincipalFromClaimsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClaims := NewMockReadableClaims(ctrl)

	mockClaims.EXPECT().Claims(gomock.Any()).Times(1).Return(fmt.Errorf("mock-error"))

	principal, err := NewPrincipalFromClaims(mockClaims)

	if principal != nil {
		t.Fatalf("returned principal should be null, but it is not")
	}

	if err == nil {
		t.Fatalf("returned error should not be null, but it is")
	}

	if err.Error() != "mock-error" {
		t.Fatalf("returned error message doesn't match expected")
	}
}

func TestPrincipalContext(t *testing.T) {
	mockCtx := context.TODO()
	mockPrincipal := &Principal{Subject: "mock-sub"}

	result := PrincipalContext(mockCtx, mockPrincipal)

	returnedPrincipal := result.Value(PrincipalContextKey).(*Principal)

	if returnedPrincipal.Subject != "mock-sub" {
		t.Fatalf("returned subject does not match expected")
	}

	result = PrincipalContext(nil, mockPrincipal)

	returnedPrincipal = result.Value(PrincipalContextKey).(*Principal)

	if returnedPrincipal.Subject != "mock-sub" {
		t.Fatalf("returned subject does not match expected")
	}
}

func TestPrincipalFromContext(t *testing.T) {
	mockCtx := context.TODO()
	mockPrincipal := &Principal{Subject: "mock-sub"}

	ctx := context.WithValue(mockCtx, PrincipalContextKey, mockPrincipal)

	principal := PrincipalFromContext(ctx)

	if principal == nil || principal.Subject != "mock-sub" {
		t.Fatalf("returned subject does not match expected")
	}
}

func mockOIDCProviderSupplier(provider *oidc.Provider, err error) OIDCProviderSupplier {
	return func(ctx context.Context, issuer string) (*oidc.Provider, error) {
		return provider, err
	}
}

func TestOAuth2Context_LoginRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)

	ctx := context.TODO()
	mockTracer.EXPECT().Start(gomock.Any(), "authentication.OAuth2Context.LoginRedirect").
		Times(1).
		Return(ctx, trace.SpanFromContext(ctx))

	config := &Config{
		Enabled:              true,
		issuer:               "http://localhost/issuer",
		clientID:             "mock-client-id",
		clientSecret:         "mock-client-secret",
		redirectURL:          "http://localhost/redirect",
		verificationStrategy: "jwks",
		scopes:               []string{"openid", "offline_access"},
	}

	oauth2Context := NewOAuth2Context(config, mockOIDCProviderSupplier(&oidc.Provider{}, nil), mockTracer, mockLogger, mockMonitor)

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/login", nil)
	mockRequest.RemoteAddr = "0.0.0.0"
	mockRequest.Header.Set("Referer", "mock-referer")

	mockResponse := httptest.NewRecorder()

	oauth2Context.LoginRedirect(mockResponse, mockRequest)

	if mockResponse.Code != http.StatusFound {
		t.Fatalf("response code error, expected %d, got %d", http.StatusFound, mockResponse.Code)
	}

	expectedLocation := "/api/v0/?audience=mock-client-id&client_id=mock-client-id&nonce=eyJyZWZlcmVyIjoibW9jay1yZWZlcmVyIiwicmVtb3RlLWFkZHJlc3MiOiIwLjAuMC4wIn0%3D&redirect_uri=http%3A%2F%2Flocalhost%2Fredirect&response_type=code&scope=openid+offline_access&state="
	location := mockResponse.Header().Get("Location")
	if !strings.HasPrefix(location, expectedLocation) {
		t.Fatalf("location header error, expected %s, got %s", expectedLocation, location)
	}
}

func TestNewOAuth2Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)

	configJWKs := Config{
		Enabled:              true,
		issuer:               "http://localhost/issuer",
		clientID:             "mock-client-id",
		clientSecret:         "mock-client-secret",
		redirectURL:          "http://localhost/redirect",
		verificationStrategy: "jwks",
		scopes:               []string{"openid", "offline_access"},
	}

	configUserinfo := configJWKs
	configUserinfo.verificationStrategy = "userinfo"

	configFailing := configUserinfo
	configFailing.verificationStrategy = "mock-error"

	for _, test := range []struct {
		name          string
		config        *Config
		supplier      OIDCProviderSupplier
		expectedError string
	}{
		{
			name:          "JWKSStrategySuccess",
			config:        &configJWKs,
			supplier:      mockOIDCProviderSupplier(&oidc.Provider{}, nil),
			expectedError: "",
		},
		{
			name:          "UserinfoStrategySuccess",
			config:        &configUserinfo,
			supplier:      mockOIDCProviderSupplier(&oidc.Provider{}, nil),
			expectedError: "",
		},
		{
			name:          "FailingStrategy",
			config:        &configFailing,
			supplier:      mockOIDCProviderSupplier(&oidc.Provider{}, nil),
			expectedError: "OAuth2VerificationStrategy value is not valid, expected one of 'jwks, userinfo', got %v",
		},
		{
			name:          "ProviderSupplierError",
			config:        &configJWKs,
			supplier:      mockOIDCProviderSupplier(&oidc.Provider{}, fmt.Errorf("mock-error")),
			expectedError: "Unable to fetch provider info, error: %v",
		},
	} {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			if tt.expectedError != "" {
				mockLogger.EXPECT().Fatalf(tt.expectedError, gomock.Eq("mock-error"))
			}

			_ = NewOAuth2Context(tt.config, tt.supplier, mockTracer, mockLogger, mockMonitor)

		})
	}
}
