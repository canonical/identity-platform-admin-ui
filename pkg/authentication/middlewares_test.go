// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestMiddleware_OAuth2Authentication(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	logger := NewMockLoggerInterface(ctrl)
	oauth2 := NewMockOAuth2ContextInterface(ctrl)

	tracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.Middleware.OAuth2Authentication")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("main handler\n"))
	})

	middleware := NewAuthenticationMiddleware(oauth2, tracer, logger)
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

			decoratedHandler := middleware.OAuth2Authentication(mainHandler)
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
