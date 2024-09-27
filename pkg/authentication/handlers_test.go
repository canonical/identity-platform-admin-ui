// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

var (
	mockTTLSeconds = 2 * 60
)

func TestHandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().Start(gomock.Any(), "authentication.OAuth2Context.LoginRedirect").
		Times(1).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	mockHelper := NewMockOAuth2HelperInterface(ctrl)
	mockHelper.EXPECT().RandomURLString().Return("mock-nonce")
	mockHelper.EXPECT().RandomURLString().Return("mock-state")

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Encrypt(gomock.Any()).Times(2).
		DoAndReturn(func(data string) (string, error) { return data, nil })

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/auth", nil)
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

	api := NewAPI(
		"",
		NewOAuth2Context(config, mockOIDCProviderSupplier(&oidc.Provider{}, nil), mockTracer, mockLogger, mockMonitor),
		mockHelper,
		NewAuthCookieManager(mockTTLSeconds, mockTTLSeconds, mockEncrypt, mockLogger),
		mockTracer,
		mockLogger,
	)

	api.handleLogin(mockResponse, mockRequest)

	if mockResponse.Code != http.StatusFound {
		t.Fatalf("response code error, expected %d, got %d", http.StatusFound, mockResponse.Code)
	}

	expectedLocation := "/api/v0/?client_id=mock-client-id&nonce=mock-nonce&redirect_uri=http%3A%2F%2Flocalhost%2Fredirect&response_type=code&scope=openid+offline_access&state=mock-state"
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

	expectedNonceValue := "mock-nonce"
	if nonceCookie == nil {
		t.Fatal("nonce cookie not found")
	}

	if nonceCookie.Value != expectedNonceValue {
		t.Fatalf("nonce cookie value does not match, expected %s, got %s", expectedNonceValue, nonceCookie.Value)
	}
}

func TestHandleLoginCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Times(1)
	mockTracer := NewMockTracer(ctrl)

	mockHelper := NewMockOAuth2HelperInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt(gomock.Any()).Times(2).
		DoAndReturn(func(data string) (string, error) { return data, nil })
	mockEncrypt.EXPECT().Encrypt(gomock.Any()).Times(3).
		DoAndReturn(func(data string) (string, error) { return data, nil })

	mockVerifier := NewMockTokenVerifier(ctrl)
	mockVerifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Return(&UserPrincipal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)

	mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
	mockOauth2Ctx.EXPECT().Verifier().Times(1).Return(mockVerifier)

	mockToken := &oauth2.Token{}
	mockToken.AccessToken = "mock-access-token"
	mockToken.RefreshToken = "mock-refresh-token"
	mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})

	mockOauth2Ctx.EXPECT().
		RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).
		Return(mockToken, nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code&state=mock-state", nil)
	mockRequest.AddCookie(&http.Cookie{
		Name:  "nonce",
		Value: "mock-nonce",
	})
	mockRequest.AddCookie(&http.Cookie{
		Name:  "state",
		Value: "mock-state",
	})

	mockResponse := httptest.NewRecorder()

	api := NewAPI(
		"",
		mockOauth2Ctx,
		mockHelper,
		NewAuthCookieManager(mockTTLSeconds, mockTTLSeconds, mockEncrypt, mockLogger),
		mockTracer,
		mockLogger,
	)

	api.handleCallback(mockResponse, mockRequest)

	result := mockResponse.Result()

	if result.StatusCode != http.StatusFound {
		t.Fatalf("response code error, expected %d, got %d", http.StatusFound, result.StatusCode)
	}

	location := result.Header.Get("Location")

	if location != ui.UIPrefix {
		t.Fatalf("redirect doesn't point to the right location, expected %s, got %s", ui.UIPrefix, location)
	}

	cookie, found := findCookie("id-token", result.Cookies())
	if !found || cookie.Value != "mock-id-token" {
		t.Fatalf("id-token cookie not found or does not match, expected %s, got %s", "mock-id-token", cookie.Value)
	}

	cookie, found = findCookie("access-token", result.Cookies())
	if !found || cookie.Value != "mock-access-token" {
		t.Fatalf("access-token cookie not found or does not match, expected %s, got %s", "mock-access-token", cookie.Value)
	}

	cookie, found = findCookie("refresh-token", result.Cookies())
	if !found || cookie.Value != "mock-refresh-token" {
		t.Fatalf("refresh-token cookie not found or does not match, expected %s, got %s", "mock-refresh-token", cookie.Value)
	}
}

func TestHandleLoginCallbackFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTracer := NewMockTracer(ctrl)

	mockToken := &oauth2.Token{}
	mockToken.AccessToken = "mock-access-token"
	mockToken.RefreshToken = "mock-refresh-token"

	mockRequestNoParams := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback", nil)
	mockRequestNoStateParam := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code", nil)

	mockRequestNoStateCookie := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code&state=mock-state", nil)

	mockRequestWithInvalidStateCookie := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code&state=mock-state", nil)
	mockRequestWithInvalidStateCookie.AddCookie(&http.Cookie{
		Name:  "state",
		Value: "invalid-state",
	})

	mockRequestWithValidStateCookie := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code&state=mock-state", nil)
	mockRequestWithValidStateCookie.AddCookie(&http.Cookie{
		Name:  "state",
		Value: "mock-state",
	})

	mockRequestWithInvalidNonce := httptest.NewRequest(http.MethodGet, "/api/v0/auth/callback?code=mock-code&state=mock-state", nil)
	mockRequestWithInvalidNonce.AddCookie(&http.Cookie{
		Name:  "state",
		Value: "mock-state",
	})
	mockRequestWithInvalidNonce.AddCookie(&http.Cookie{
		Name:  "nonce",
		Value: "invalid-nonce",
	})

	for _, tt := range []struct {
		name         string
		errorMessage string
		request      *http.Request
		setupMocks   func(*MockOAuth2ContextInterface, *MockLoggerInterface, *MockTokenVerifier, *MockEncryptInterface)
	}{
		{
			name:    "CodeParamNotFound",
			request: mockRequestNoParams,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Error("OAuth2 code not found")
			},
			errorMessage: "OAuth2 code not found",
		},
		{
			name:    "StateParamNotFound",
			request: mockRequestNoStateParam,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Error("OAuth2 state not found")
			},
			errorMessage: "OAuth2 state not found",
		},
		{
			name:    "StateCookieNotFound",
			request: mockRequestNoStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Error("state cookie not found")
			},
			errorMessage: "state cookie not found",
		},
		{
			name:    "StateCookieNotValid",
			request: mockRequestWithInvalidStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Error("state parameter does not match state cookie")
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(1).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "state parameter does not match state cookie",
		},
		{
			name:    "RetrieveTokenError",
			request: mockRequestWithValidStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Debugf("user login second leg with code '%s'", "mock-code").Times(1)
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Times(1).Return(nil, errors.New("mock-error"))
				logger.EXPECT().Errorf("unable to retrieve tokens with code '%s', error: %v", "mock-code", errors.New("mock-error"))
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(1).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "mock-error",
		},
		{
			name:    "IDTokenNotFound",
			request: mockRequestWithValidStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Debugf("user login second leg with code '%s'", "mock-code").Times(1)
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)
				logger.EXPECT().Error("unable to retrieve ID token")
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(1).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "unable to retrieve ID token",
		},
		{
			name:    "IDTokenNotVerifiable",
			request: mockRequestWithValidStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Debugf("user login second leg with code '%s'", "mock-code").Times(1)
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("mock-error"))
				oauth2Ctx.EXPECT().Verifier().Return(verifier)

				logger.EXPECT().Errorf("unable to verify ID token, error: %v", errors.New("mock-error"))
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(1).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "mock-error",
		},
		{
			name:    "NonceCookieNotFound",
			request: mockRequestWithValidStateCookie,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Debugf("user login second leg with code '%s'", "mock-code").Times(1)
				logger.EXPECT().Error("nonce cookie not found")
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).
					Return(&UserPrincipal{
						Subject: "mock-subject",
						Nonce:   "mock-nonce",
					}, nil)

				oauth2Ctx.EXPECT().Verifier().Return(verifier)
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(1).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "nonce cookie not found",
		},
		{
			name:    "NonceCookieNotValid",
			request: mockRequestWithInvalidNonce,
			setupMocks: func(oauth2Ctx *MockOAuth2ContextInterface, logger *MockLoggerInterface, verifier *MockTokenVerifier, encrypt *MockEncryptInterface) {
				logger.EXPECT().Debugf("user login second leg with code '%s'", "mock-code").Times(1)
				logger.EXPECT().Error("id token nonce does not match nonce cookie")
				mockToken = mockToken.WithExtra(map[string]interface{}{"id_token": "mock-id-token"})
				oauth2Ctx.EXPECT().RetrieveTokens(gomock.Any(), gomock.Eq("mock-code")).Return(mockToken, nil)

				verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Times(1).
					Return(&UserPrincipal{
						Subject: "mock-subject",
						Nonce:   "mock-nonce",
					}, nil)

				oauth2Ctx.EXPECT().Verifier().Return(verifier)
				encrypt.EXPECT().Decrypt(gomock.Any()).Times(2).DoAndReturn(func(s string) (string, error) { return s, nil })
			},
			errorMessage: "id token nonce does not match nonce cookie",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
			mockVerifier := NewMockTokenVerifier(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)

			mockHelper := NewMockOAuth2HelperInterface(ctrl)
			mockEncrypt := NewMockEncryptInterface(ctrl)

			tt.setupMocks(mockOauth2Ctx, mockLogger, mockVerifier, mockEncrypt)

			mockResponse := httptest.NewRecorder()

			api := NewAPI(
				"",
				mockOauth2Ctx,
				mockHelper,
				NewAuthCookieManager(mockTTLSeconds, mockTTLSeconds, mockEncrypt, mockLogger),
				mockTracer,
				mockLogger,
			)
			api.handleCallback(mockResponse, tt.request)

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

func TestHandleMe(t *testing.T) {
	ctrl := gomock.NewController(t)
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

	mockTracer := NewMockTracer(ctrl)
	mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)

	mockHelper := NewMockOAuth2HelperInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockResponse := httptest.NewRecorder()

	api := NewAPI(
		"",
		mockOauth2Ctx,
		mockHelper,
		NewAuthCookieManager(mockTTLSeconds, mockTTLSeconds, mockEncrypt, mockLogger),
		mockTracer,
		mockLogger,
	)

	mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/auth/me", nil)
	mockCtx := PrincipalContext(context.Background(), p)
	mockRequest = mockRequest.WithContext(mockCtx)

	api.handleMe(mockResponse, mockRequest)

	response := mockResponse.Result()
	defer response.Body.Close()

	got := new(UserPrincipal)
	_ = json.NewDecoder(response.Body).Decode(got)

	expectedPrincipal := UserPrincipal{
		Subject:         "mock-subject",
		Name:            "mock-name",
		Email:           "mock-email",
		SessionID:       "mock-sid",
		Nonce:           "mock-nonce",
		RawAccessToken:  "",
		RawIdToken:      "",
		RawRefreshToken: "",
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("response object status error, expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	if !reflect.DeepEqual(*got, expectedPrincipal) {
		t.Fatalf("response body not matching, expected %v, got %v", expectedPrincipal, *got)
	}
}

func TestLogout(t *testing.T) {
	ctrl := gomock.NewController(t)

	for _, tt := range []struct {
		name         string
		errorMessage string
		setupMocks   func(*MockOAuth2ContextInterface, *MockAuthCookieManagerInterface, *MockLoggerInterface)
	}{
		{
			name:         "Success",
			errorMessage: "",
			setupMocks: func(c *MockOAuth2ContextInterface, m *MockAuthCookieManagerInterface, l *MockLoggerInterface) {
				m.EXPECT().ClearIDTokenCookie(gomock.Any())
				m.EXPECT().ClearAccessTokenCookie(gomock.Any())
				m.EXPECT().ClearRefreshTokenCookie(gomock.Any())

				c.EXPECT().Logout(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
		},
		{
			name:         "Failure",
			errorMessage: "logout request failed, err mock-err",
			setupMocks: func(c *MockOAuth2ContextInterface, m *MockAuthCookieManagerInterface, l *MockLoggerInterface) {
				err := errors.New("mock-error")
				l.EXPECT().Errorf("logout request failed, err %v", err)

				c.EXPECT().Logout(gomock.Any(), gomock.Any()).Times(1).Return(err)

				m.EXPECT().ClearNonceCookie(gomock.Any()).Times(1)
				m.EXPECT().ClearStateCookie(gomock.Any()).Times(1)
				m.EXPECT().ClearNextToCookie(gomock.Any()).Times(1)
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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

			mockTracer := NewMockTracer(ctrl)
			mockOauth2Ctx := NewMockOAuth2ContextInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)

			mockHelper := NewMockOAuth2HelperInterface(ctrl)

			mockResponse := httptest.NewRecorder()

			mockCookieManager := NewMockAuthCookieManagerInterface(ctrl)
			api := NewAPI(
				"",
				mockOauth2Ctx,
				mockHelper,
				mockCookieManager,
				mockTracer,
				mockLogger,
			)

			mockRequest := httptest.NewRequest(http.MethodGet, "/api/v0/auth/logout", nil)
			mockCtx := PrincipalContext(context.Background(), p)
			mockRequest = mockRequest.WithContext(mockCtx)

			tt.setupMocks(mockOauth2Ctx, mockCookieManager, mockLogger)

			api.handleLogout(mockResponse, mockRequest)

			response := mockResponse.Result()

			if tt.errorMessage != "" {
				if response.StatusCode != http.StatusBadRequest {
					t.Fatalf("response code error, expected %d, got %d", http.StatusBadRequest, response.StatusCode)
				}

			} else {
				if response.StatusCode != http.StatusFound {
					t.Fatalf("response code error, expected %d, got %d", http.StatusBadRequest, response.StatusCode)
				}
			}

		})
	}
}
