// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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

func TestMiddleware_OAuth2AuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	logger := NewMockLoggerInterface(ctrl)

	verifier := NewMockTokenVerifier(ctrl)
	verifier.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock-error"))
	verifier.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).Return(&ServicePrincipal{Subject: "mock-subject"}, nil)

	oauth2 := NewMockOAuth2ContextInterface(ctrl)
	oauth2.EXPECT().Verifier().Times(2).Return(verifier)

	cookieManager := NewMockAuthCookieManagerInterface(ctrl)
	cookieManager.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
	cookieManager.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
	cookieManager.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

	cookieManager.EXPECT().ClearIDTokenCookie(gomock.Any())
	cookieManager.EXPECT().ClearAccessTokenCookie(gomock.Any())
	cookieManager.EXPECT().ClearRefreshTokenCookie(gomock.Any())

	tracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	tracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("main handler\n"))
	})

	middleware := NewAuthenticationMiddleware(oauth2, cookieManager, tracer, logger)
	middleware.SetAllowListedEndpoints("/api/v0/different-mock-key")

	for _, tt := range []struct {
		name                 string
		expctedResponse      string
		expctedErrorResponse *types.Response
		expectedCode         int
		mockRequest          *http.Request
	}{
		{
			name:         "ProtectedEndpoint",
			expectedCode: http.StatusUnauthorized,
			mockRequest:  httptest.NewRequest(http.MethodGet, "/api/v0/mock-key", nil),
		},
		{
			name:            "AllowelistedEndpoint",
			expctedResponse: "main handler\n",
			expectedCode:    http.StatusOK,
			mockRequest:     httptest.NewRequest(http.MethodGet, "/api/v0/different-mock-key", nil),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockResponse := httptest.NewRecorder()

			decoratedHandler := applyMiddlewares(mainHandler, middleware.OAuth2AuthenticationChain()...)
			decoratedHandler.ServeHTTP(mockResponse, tt.mockRequest)

			result := mockResponse.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.expectedCode {
				t.Fatalf("actual response code differes from expected")
			}

			if tt.expctedResponse != "" {
				if mockResponse.Body.String() != tt.expctedResponse {
					t.Fatalf("actual response and expected response do not match")
				}

				return
			}

			response := new(types.Response)
			_ = json.NewDecoder(result.Body).Decode(response)

			if response.Status != tt.expectedCode || result.StatusCode != tt.expectedCode {
				t.Fatalf("actual response code differes from expected")
			}

			if !strings.HasPrefix(response.Message, "unauthorized") {
				t.Fatalf("actual response body differes from expected")
			}
		})
	}
}

func TestMiddleware_OAuth2AuthenticationMiddlewareFailures(t *testing.T) {
	ctrl := gomock.NewController(t)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("main handler\n"))
	})

	mockRequestNoBearer := httptest.NewRequest(http.MethodGet, "/api/v0/protected", nil)

	mockReqWithBearer := httptest.NewRequest(http.MethodGet, "/api/v0/protected", nil)
	mockReqWithBearer.Header.Set("Authorization", "Bearer mock-access-token")

	tests := []struct {
		name       string
		setupMocks func(*MockAuthCookieManagerInterface, *MockTokenVerifier, *MockOAuth2ContextInterface, *MockLoggerInterface, *MockTracer)
		request    *http.Request
		expected   string
	}{
		{
			name:    "NoAuth",
			request: mockRequestNoBearer,
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("")

				c.EXPECT().ClearIDTokenCookie(gomock.Any())
				c.EXPECT().ClearAccessTokenCookie(gomock.Any())
				c.EXPECT().ClearRefreshTokenCookie(gomock.Any())
			},
			expected: "unauthorized: unable to authenticate from either bearer or cookie token, no authentication token found",
		},
		{
			name:    "BearerAuthFailure",
			request: mockReqWithBearer,
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				v.EXPECT().VerifyAccessToken(gomock.Any(), "mock-access-token").
					Return(nil, errors.New("mock-error"))

				o.EXPECT().Verifier().Times(1).Return(v)

				c.EXPECT().ClearIDTokenCookie(gomock.Any())
				c.EXPECT().ClearAccessTokenCookie(gomock.Any())
				c.EXPECT().ClearRefreshTokenCookie(gomock.Any())

				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			},
			expected: "unauthorized: mock-error",
		},
		{
			name:    "CookieAuthInvalidIDToken",
			request: mockRequestNoBearer,
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))

				v.EXPECT().VerifyIDToken(gomock.Any(), "mock-id-token").
					Times(1).
					Return(nil, errors.New("mock-error"))
				v.EXPECT().VerifyAccessToken(gomock.Any(), "mock-access-token").
					Times(1).
					Return(&ServicePrincipal{Subject: "mock-subject"}, nil)

				o.EXPECT().Verifier().Times(2).Return(v)

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

				c.EXPECT().ClearIDTokenCookie(gomock.Any())
				c.EXPECT().ClearAccessTokenCookie(gomock.Any())
				c.EXPECT().ClearRefreshTokenCookie(gomock.Any())
			},
			expected: "unauthorized: unable to authenticate from either bearer or cookie token, mock-error",
		},
		{
			name:    "CookieAuthInvalidAccessToken",
			request: mockRequestNoBearer,
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))

				v.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&UserPrincipal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)
				v.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, errors.New("mock-error"))

				o.EXPECT().Verifier().Times(2).Return(v)

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

				c.EXPECT().ClearIDTokenCookie(gomock.Any())
				c.EXPECT().ClearAccessTokenCookie(gomock.Any())
				c.EXPECT().ClearRefreshTokenCookie(gomock.Any())
			},
			expected: "unauthorized: unable to authenticate from either bearer or cookie token, mock-error",
		},
		{
			name:    "CookieAuthRefreshError",
			request: mockRequestNoBearer,
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))

				v.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, &oidc.TokenExpiredError{})
				v.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&UserPrincipal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)

				o.EXPECT().Verifier().Times(2).Return(v)

				token := &oauth2.Token{
					AccessToken:  "mock-access-token",
					TokenType:    "bearer",
					RefreshToken: "mock-refresh-token",
				}
				token = token.WithExtra(map[string]any{"id_token": "mock-id-token"})

				o.EXPECT().RefreshToken(gomock.Any(), "mock-refresh-token").
					Times(1).
					Return(nil, errors.New("mock-refresh-error"))

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

				c.EXPECT().ClearIDTokenCookie(gomock.Any())
				c.EXPECT().ClearAccessTokenCookie(gomock.Any())
				c.EXPECT().ClearRefreshTokenCookie(gomock.Any())
			},
			expected: "unauthorized: unable to authenticate from either bearer or cookie token, mock-refresh-error",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tracer := NewMockTracer(ctrl)
			logger := NewMockLoggerInterface(ctrl)

			verifier := NewMockTokenVerifier(ctrl)
			oauth2 := NewMockOAuth2ContextInterface(ctrl)

			cookieManager := NewMockAuthCookieManagerInterface(ctrl)

			tt.setupMocks(cookieManager, verifier, oauth2, logger, tracer)

			middleware := NewAuthenticationMiddleware(oauth2, cookieManager, tracer, logger)
			m := applyMiddlewares(mainHandler, middleware.OAuth2AuthenticationChain()...)

			mockResponse := httptest.NewRecorder()

			m.ServeHTTP(mockResponse, tt.request)

			response := mockResponse.Result()

			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status does not match, exptected %d, got %d", http.StatusUnauthorized, response.StatusCode)
			}

			body := response.Body
			defer body.Close()

			respObj := new(types.Response)
			_ = json.NewDecoder(body).Decode(respObj)

			if respObj.Status != http.StatusUnauthorized {
				t.Fatalf("expected status does not match, exptected %d, got %d", http.StatusUnauthorized, respObj.Status)
			}

			if respObj.Message != tt.expected {
				t.Fatalf("expected error message does not match, expected %s, got %s", tt.expected, respObj.Message)
			}
		})
	}
}

func TestMiddleware_OAuth2AuthenticationMiddlewareSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("main handler\n"))
	})

	mockReqWithBearer := httptest.NewRequest(http.MethodGet, "/api/v0/protected", nil)
	mockReqWithBearer.Header.Set("Authorization", "Bearer mock-access-token")

	mockReqWithCookies := httptest.NewRequest(http.MethodGet, "/api/v0/protected", nil)

	mockReqWithCookiesRefresh := httptest.NewRequest(http.MethodGet, "/api/v0/protected", nil)

	tests := []struct {
		name       string
		setupMocks func(*MockAuthCookieManagerInterface, *MockTokenVerifier, *MockOAuth2ContextInterface, *MockLoggerInterface, *MockTracer)
		request    *http.Request
	}{
		{
			name: "BearerAuthSuccess",
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				v.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&ServicePrincipal{Subject: "mock-subject"}, nil)

				o.EXPECT().Verifier().Times(1).Return(v)

				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			},
			request: mockReqWithBearer,
		},
		{
			name: "CookieAuthSuccess",
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				v.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&UserPrincipal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)
				v.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&ServicePrincipal{Subject: "mock-subject"}, nil)

				o.EXPECT().Verifier().Times(2).Return(v)

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			},
			request: mockReqWithCookies,
		},
		{
			name: "CookieAuthRefreshSuccess",
			setupMocks: func(c *MockAuthCookieManagerInterface, v *MockTokenVerifier, o *MockOAuth2ContextInterface, l *MockLoggerInterface, t *MockTracer) {
				v.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				v.EXPECT().VerifyAccessToken(gomock.Any(), gomock.Any()).
					Return(nil, &oidc.TokenExpiredError{})

				o.EXPECT().Verifier().Times(3).Return(v)
				token := &oauth2.Token{
					AccessToken:  "mock-access-token",
					TokenType:    "bearer",
					RefreshToken: "mock-refresh-token",
				}
				token = token.WithExtra(map[string]any{"id_token": "mock-id-token"})

				o.EXPECT().RefreshToken(gomock.Any(), "mock-refresh-token").
					Times(1).
					Return(token, nil)

				v.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Return(&UserPrincipal{Subject: "mock-subject", Nonce: "mock-nonce"}, nil)

				c.EXPECT().GetIDTokenCookie(gomock.Any()).Return("mock-id-token")
				c.EXPECT().GetAccessTokenCookie(gomock.Any()).Return("mock-access-token")
				c.EXPECT().GetRefreshTokenCookie(gomock.Any()).Return("mock-refresh-token")

				c.EXPECT().SetIDTokenCookie(gomock.Any(), "mock-id-token")
				c.EXPECT().SetAccessTokenCookie(gomock.Any(), "mock-access-token")
				c.EXPECT().SetRefreshTokenCookie(gomock.Any(), "mock-refresh-token")

				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2BearerAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
				t.EXPECT().
					Start(gomock.Any(), gomock.Eq("authentication.Middleware.oAuth2CookieAuthentication")).
					Times(1).
					Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			},
			request: mockReqWithCookiesRefresh,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tracer := NewMockTracer(ctrl)
			logger := NewMockLoggerInterface(ctrl)

			verifier := NewMockTokenVerifier(ctrl)
			oauth2Ctx := NewMockOAuth2ContextInterface(ctrl)

			cookieManager := NewMockAuthCookieManagerInterface(ctrl)

			tt.setupMocks(cookieManager, verifier, oauth2Ctx, logger, tracer)

			middleware := NewAuthenticationMiddleware(oauth2Ctx, cookieManager, tracer, logger)
			m := applyMiddlewares(mainHandler, middleware.OAuth2AuthenticationChain()...)

			mockResponse := httptest.NewRecorder()

			m.ServeHTTP(mockResponse, tt.request)

			response := mockResponse.Result()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected status does not match, exptected %d, got %d", http.StatusOK, response.StatusCode)
			}

			body := response.Body
			defer body.Close()

			stringBytes, _ := io.ReadAll(body)

			expected := "main handler\n"
			bodyString := string(stringBytes)

			if expected != bodyString {
				t.Fatalf("expected message does not match, expected %s, got %s", bodyString, expected)
			}
		})
	}
}

func applyMiddlewares(handler http.Handler, ms ...func(http.Handler) http.Handler) http.Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler
}
