// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestHandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().Start(gomock.Any(), "authentication.OAuth2Context.LoginRedirect").
		Times(1).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/login", nil)
	mockRequest.RemoteAddr = "0.0.0.0"
	mockRequest.Header.Set("Referer", "mock-referer")
	mockResponse := httptest.NewRecorder()

	config := &Config{
		Enabled:              true,
		issuer:               "http://localhost/issuer",
		clientID:             "mock-client-id",
		clientSecret:         "mock-client-secret",
		redirectURL:          "http://localhost/redirect",
		verificationStrategy: "jwks",
		scopes:               []string{"openid", "offline_access"},
	}
	api := NewAPI(NewOAuth2Context(config, mockOIDCProviderSupplier(&oidc.Provider{}, nil), mockTracer, mockLogger, mockMonitor), mockTracer, mockLogger)

	api.handleLogin(mockResponse, mockRequest)

	if mockResponse.Code != http.StatusFound {
		t.Fatalf("response code error, expected %d, got %d", http.StatusFound, mockResponse.Code)
	}

	expectedLocation := "/api/v0/?audience=mock-client-id&client_id=mock-client-id&nonce=eyJyZWZlcmVyIjoibW9jay1yZWZlcmVyIiwicmVtb3RlLWFkZHJlc3MiOiIwLjAuMC4wIn0%3D&redirect_uri=http%3A%2F%2Flocalhost%2Fredirect&response_type=code&scope=openid+offline_access&state="
	location := mockResponse.Header().Get("Location")
	if !strings.HasPrefix(location, expectedLocation) {
		t.Fatalf("location header error, expected %s, got %s", expectedLocation, location)
	}

	response := mockResponse.Result()
	var nonceCookie *http.Cookie = nil
	for _, cookie := range response.Cookies() {
		if cookie.Name == "nonce" {
			nonceCookie = cookie
		}
	}

	expectedNonceValue := "eyJyZWZlcmVyIjoibW9jay1yZWZlcmVyIiwicmVtb3RlLWFkZHJlc3MiOiIwLjAuMC4wIn0="
	if nonceCookie == nil {
		t.Fatal("nonce cookie not found")
	}

	if nonceCookie.Value != expectedNonceValue {
		t.Fatalf("nonce cookie value does not match, expected %s, got %s", expectedNonceValue, nonceCookie.Value)
	}
}

func TestHandleLoginWithCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Times(1)
	mockTracer := NewMockTracer(ctrl)

	mockVerifier := NewMockTokenVerifier(ctrl)
	mockVerifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Return(&Principal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)

	mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
	mockOauth2Ctx.EXPECT().Verifier().Times(1).Return(mockVerifier)

	mockToken := &oauth2.Token{}
	mockToken.AccessToken = "mock-access-token"
	mockToken.RefreshToken = "mock-refresh-token"
	mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})

	mockOauth2Ctx.EXPECT().
		RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).
		Return(mockToken, nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/login?code=mock-code", nil)
	mockRequest.AddCookie(&http.Cookie{
		Name:  "nonce",
		Value: "mock-nonce",
	})
	mockResponse := httptest.NewRecorder()

	api := NewAPI(mockOauth2Ctx, mockTracer, mockLogger)

	api.handleLogin(mockResponse, mockRequest)

	result := mockResponse.Result()

	if result.StatusCode != http.StatusOK {
		t.Fatalf("response code error, expected %d, got %d", http.StatusOK, result.StatusCode)
	}

	body := result.Body
	defer result.Body.Close()

	tokens := new(oauth2Tokens)

	_ = json.NewDecoder(body).Decode(tokens)

	if tokens.AccessToken != "mock-access-token" {
		t.Fatalf("access token does not match expected, got %s, expected %s", tokens.AccessToken, "mock-access-token")
	}

	if tokens.RefreshToken != "mock-refresh-token" {
		t.Fatalf("refresh token does not match expected, got %s, expected %s", tokens.RefreshToken, "mock-refresh-token")
	}

	if tokens.IDToken != "mock-id-token" {
		t.Fatalf("id token does not match expected, got %s, expected %s", tokens.IDToken, "mock-id-token")
	}
}

func TestHandleLoginWithCodeFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTracer := NewMockTracer(ctrl)

	mockToken := &oauth2.Token{}
	mockToken.AccessToken = "mock-access-token"
	mockToken.RefreshToken = "mock-refresh-token"

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/login?code=mock-code", nil)
	mockRequestWithNonce := httptest.NewRequest(http.MethodGet, "/api/v0/login?code=mock-code", nil)
	mockRequestWithNonce.AddCookie(&http.Cookie{
		Name:  "nonce",
		Value: "invalid-nonce",
	})

	for _, tt := range []struct {
		name         string
		errorMessage string
		request      *http.Request
		setupMocks   func(*MockOAuth2ContextInterface, *MockLoggerInterface, *MockTokenVerifier)
	}{
		{
			name:    "RetrieveTokenError",
			request: mockRequest,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier) {
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Times(1).Return(nil, errors.New("mock-error"))
				logger.EXPECT().Errorf("unable to retrieve tokens with code '%s', error: %v", "mock-code", errors.New("mock-error"))
			},
			errorMessage: "mock-error",
		},
		{
			name:    "IDTokenNotFound",
			request: mockRequest,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier) {
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)
				logger.EXPECT().Error("unable to retrieve ID token")
			},
			errorMessage: "unable to retrieve ID token",
		},
		{
			name:    "IDTokenNotVerifiable",
			request: mockRequest,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier) {
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("mock-error"))
				oauth2Ctx.EXPECT().Verifier().Return(verifier)

				logger.EXPECT().Errorf("unable to verify ID token, error: %v", errors.New("mock-error"))
			},
			errorMessage: "mock-error",
		},
		{
			name:    "NonceCookieNotFound",
			request: mockRequest,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier) {
				logger.EXPECT().Error("nonce cookie not found")
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).
					Return(&Principal{
						Subject: "mock-subject",
						Nonce:   "mock-nonce",
					}, nil)

				oauth2Ctx.EXPECT().Verifier().Return(verifier)
			},
			errorMessage: "nonce cookie not found",
		},
		{
			name:    "NonceCookieNotValid",
			request: mockRequestWithNonce,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier) {
				logger.EXPECT().Error("id token nonce does not match")
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).
					Return(&Principal{
						Subject: "mock-subject",
						Nonce:   "mock-nonce",
					}, nil)

				oauth2Ctx.EXPECT().Verifier().Return(verifier)
			},
			errorMessage: "id token nonce error",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
			mockVerifier := NewMockTokenVerifier(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Times(1)

			tt.setupMocks(mockOauth2Ctx, mockLogger, mockVerifier)

			mockResponse := httptest.NewRecorder()

			api := NewAPI(mockOauth2Ctx, mockTracer, mockLogger)
			api.handleLogin(mockResponse, tt.request)

			result := mockResponse.Result()

			if result.StatusCode != http.StatusBadRequest {
				t.Fatalf("response code error, expected %d, got %d", http.StatusBadRequest, result.StatusCode)
			}

			body := result.Body
			defer result.Body.Close()

			response := new(types.Response)

			err := json.NewDecoder(body).Decode(response)
			_ = err
			if response.Status != http.StatusBadRequest {
				t.Fatalf("response object status error, expected %d, got %d", http.StatusBadRequest, response.Status)
			}

			if response.Message != tt.errorMessage {
				t.Fatalf("response message error, expected %s, got %s", tt.errorMessage, response.Message)
			}
		})
	}
}
