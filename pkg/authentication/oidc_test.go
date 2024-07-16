// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	client "github.com/ory/hydra-client-go/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_hydra.go github.com/ory/hydra-client-go/v2 OAuth2Api
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_clients.go -source=../clients/interfaces.go

func TestNewPrincipalFromClaims(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClaims := NewMockReadableClaims(ctrl)

	mockClaims.EXPECT().Claims(gomock.Any()).Times(1).
		DoAndReturn(
			func(p *UserPrincipal) error {
				p.Subject = "mock-sub"
				return nil
			})

	principal, err := NewUserPrincipalFromClaims(mockClaims)

	if err != nil {
		t.Fatalf("returned error should be null, but it is not")
	}

	if principal.Identifier() != "mock-sub" {
		t.Fatalf("returned principal subject doesn't match expected")
	}
}

func TestNewPrincipalFromClaimsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClaims := NewMockReadableClaims(ctrl)

	mockClaims.EXPECT().Claims(gomock.Any()).Times(1).Return(fmt.Errorf("mock-error"))

	principal, err := NewUserPrincipalFromClaims(mockClaims)

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
	mockPrincipal := &UserPrincipal{Subject: "mock-sub"}

	result := PrincipalContext(mockCtx, mockPrincipal)

	returnedPrincipal := result.Value(PrincipalContextKey).(*UserPrincipal)

	if returnedPrincipal.Subject != "mock-sub" {
		t.Fatalf("returned subject does not match expected")
	}

	result = PrincipalContext(nil, mockPrincipal)

	returnedPrincipal = result.Value(PrincipalContextKey).(*UserPrincipal)

	if returnedPrincipal.Subject != "mock-sub" {
		t.Fatalf("returned subject does not match expected")
	}
}

func TestPrincipalFromContext(t *testing.T) {
	mockCtx := context.TODO()
	mockPrincipal := &UserPrincipal{Subject: "mock-sub"}

	ctx := context.WithValue(mockCtx, PrincipalContextKey, mockPrincipal)

	principal := PrincipalFromContext(ctx)

	if principal == nil || principal.Identifier() != "mock-sub" {
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
		redirectURL:          "http://localhost/api/v0/auth/callback",
		verificationStrategy: "jwks",
		scopes:               []string{"openid", "offline_access"},
	}

	oauth2Context := NewOAuth2Context(config, mockOIDCProviderSupplier(&oidc.Provider{}, nil), mockTracer, mockLogger, mockMonitor)
	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/auth", nil)

	location := oauth2Context.LoginRedirect(mockRequest.Context(), "mock-nonce", "mock-state")

	expectedLocation := "?audience=mock-client-id&client_id=mock-client-id&nonce=mock-nonce&redirect_uri=http%3A%2F%2Flocalhost%2Fapi%2Fv0%2Fauth%2Fcallback&response_type=code&scope=openid+offline_access&state=mock-state"

	if location != expectedLocation {
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

func TestOAuth2Context_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := &UserPrincipal{
		Subject:         "mock-subject",
		Name:            "mock-name",
		Email:           "mock-email",
		SessionID:       "mock-sid",
		Nonce:           "mock-nonce",
		RawAccessToken:  "mock-access-token",
		RawIdToken:      "mock-id-token",
		RawRefreshToken: "mock-refresh-token",
	}

	errorStructString := "{\"error\":\"mock-error\", \"error_description\":\"mock-error-descr\"}"
	tests := []struct {
		name          string
		principal     *UserPrincipal
		setupMock     func(*MockOAuth2Api, *MockOAuth2Api)
		expectedError string
	}{
		{
			name:      "Success",
			principal: p,
			setupMock: func(admin, public *MockOAuth2Api) {
				sessionRevokeRequest := client.OAuth2ApiRevokeOAuth2LoginSessionsRequest{ApiService: admin}
				admin.EXPECT().RevokeOAuth2LoginSessions(gomock.Any()).Return(sessionRevokeRequest)

				admin.EXPECT().RevokeOAuth2LoginSessionsExecute(gomock.Any()).
					Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

				tokenRevokeReq := client.OAuth2ApiRevokeOAuth2TokenRequest{ApiService: public}
				public.EXPECT().RevokeOAuth2Token(gomock.Any()).Return(tokenRevokeReq)

				public.EXPECT().RevokeOAuth2TokenExecute(gomock.Any()).
					Return(&http.Response{StatusCode: http.StatusOK}, nil)
			},
			expectedError: "",
		},
		{
			name:          "NoPrincipal",
			setupMock:     func(admin, public *MockOAuth2Api) {},
			expectedError: "no principal provided",
		},
		{
			name:      "RevokeSessionFailure",
			principal: p,
			setupMock: func(admin, public *MockOAuth2Api) {
				sessionRevokeRequest := client.OAuth2ApiRevokeOAuth2LoginSessionsRequest{ApiService: admin}
				admin.EXPECT().RevokeOAuth2LoginSessions(gomock.Any()).Return(sessionRevokeRequest)
				admin.EXPECT().RevokeOAuth2LoginSessionsExecute(gomock.Any()).
					Return(nil, errors.New("mock-error"))
			},
			expectedError: "mock-error",
		},
		{
			name:      "RevokeSessionStatusCodeFailure",
			principal: p,
			setupMock: func(admin, public *MockOAuth2Api) {
				sessionRevokeRequest := client.OAuth2ApiRevokeOAuth2LoginSessionsRequest{ApiService: admin}
				admin.EXPECT().RevokeOAuth2LoginSessions(gomock.Any()).Return(sessionRevokeRequest)
				admin.EXPECT().RevokeOAuth2LoginSessionsExecute(gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       io.NopCloser(strings.NewReader(errorStructString)),
						},
						nil,
					)

			},
			expectedError: "revoke session request failed, error: mock-error, description: mock-error-descr",
		},
		{
			name:      "RevokeTokenFailure",
			principal: p,
			setupMock: func(admin, public *MockOAuth2Api) {
				sessionRevokeRequest := client.OAuth2ApiRevokeOAuth2LoginSessionsRequest{ApiService: admin}
				admin.EXPECT().RevokeOAuth2LoginSessions(gomock.Any()).
					Return(sessionRevokeRequest)
				admin.EXPECT().RevokeOAuth2LoginSessionsExecute(gomock.Any()).
					Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

				tokenRevokeReq := client.OAuth2ApiRevokeOAuth2TokenRequest{ApiService: public}
				public.EXPECT().RevokeOAuth2Token(gomock.Any()).
					Return(tokenRevokeReq)
				public.EXPECT().RevokeOAuth2TokenExecute(gomock.Any()).
					Return(nil, errors.New("mock-error"))
			},
			expectedError: "mock-error",
		},
		{
			name:      "RevokeTokenStatusCodeFailure",
			principal: p,
			setupMock: func(admin, public *MockOAuth2Api) {
				sessionRevokeRequest := client.OAuth2ApiRevokeOAuth2LoginSessionsRequest{ApiService: admin}
				admin.EXPECT().RevokeOAuth2LoginSessions(gomock.Any()).
					Return(sessionRevokeRequest)
				admin.EXPECT().RevokeOAuth2LoginSessionsExecute(gomock.Any()).
					Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

				tokenRevokeReq := client.OAuth2ApiRevokeOAuth2TokenRequest{ApiService: public}
				public.EXPECT().RevokeOAuth2Token(gomock.Any()).
					Return(tokenRevokeReq)
				public.EXPECT().RevokeOAuth2TokenExecute(gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       io.NopCloser(strings.NewReader(errorStructString)),
						},
						nil,
					)
			},
			expectedError: "revoke token request failed, error: mock-error, description: mock-error-descr",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)

			mockTracer.EXPECT().Start(gomock.Any(), "authentication.OAuth2Context.Logout").
				Times(1).
				Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockHydraAdminClient := NewMockOAuth2Api(ctrl)
			mockHydraPublicClient := NewMockOAuth2Api(ctrl)

			mockInternalHydraAdminClient := NewMockHydraClientInterface(ctrl)
			mockInternalHydraAdminClient.EXPECT().OAuth2Api().AnyTimes().Return(mockHydraAdminClient)
			mockInternalHydraPublicClient := NewMockHydraClientInterface(ctrl)
			mockInternalHydraPublicClient.EXPECT().OAuth2Api().AnyTimes().Return(mockHydraPublicClient)

			config := &Config{
				Enabled:              true,
				issuer:               "http://localhost/issuer",
				clientID:             "mock-client-id",
				clientSecret:         "mock-client-secret",
				redirectURL:          "http://localhost/api/v0/auth/callback",
				verificationStrategy: "jwks",
				scopes:               []string{"openid", "offline_access"},
				hydraPublicAPIClient: mockInternalHydraPublicClient,
				hydraAdminAPIClient:  mockInternalHydraAdminClient,
			}

			tt.setupMock(mockHydraAdminClient, mockHydraPublicClient)

			oauth2Context := NewOAuth2Context(config, mockOIDCProviderSupplier(&oidc.Provider{}, nil), mockTracer, mockLogger, mockMonitor)

			result := oauth2Context.Logout(context.TODO(), tt.principal)

			if tt.expectedError == "" {
				if result != nil {
					t.Fatalf("unexpected error returned, expected nil, got %v", result)
				}
			}

			if tt.expectedError != "" {
				if result == nil {
					t.Fatalf("unexpected nil value returned, expected %v", tt.expectedError)
				}

				errorMessage := result.Error()
				if errorMessage != tt.expectedError {
					t.Fatalf("expected error message does not match, expected %v, got %v", tt.expectedError, errorMessage)
				}
			}

		})
	}
}
