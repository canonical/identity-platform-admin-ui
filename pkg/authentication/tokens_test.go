// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestJWKSTokenVerifier_VerifyAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.JWKSTokenVerifier.VerifyAccessToken")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	mockToken := new(oidc.IDToken)
	mockToken.Subject = "mock-subject"

	for _, tt := range []struct {
		name  string
		token string
	}{
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJtb2NrLWNsaWVudC1pZCIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.BdspASNsnxeXnqZXZnFnkvv-ClMq0U6X1gCIUrh9V7c",
		},
		{
			name:  "Fail",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJkaWZmZXJlbnQtY2xpZW50LWlkIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.ZC8qlDOhxOZyUTiAk2ICpEmSnmEb2UHXAJyOrrYVCDc",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockProviderInterface(ctrl)
			tokenVerifier := oidc.NewVerifier("", nil, &oidc.Config{
				ClientID:                   "mock-client-id",
				SkipExpiryCheck:            true,
				SkipIssuerCheck:            true,
				InsecureSkipSignatureCheck: true,
			})
			mockProvider.EXPECT().Verifier(&oidc.Config{ClientID: "mock-client-id"}).Return(tokenVerifier)

			verifier := NewJWKSTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor)

			token, err := verifier.VerifyAccessToken(context.TODO(), tt.token)

			if tt.name == "Fail" && (err == nil || err.Error() != "oidc: expected audience \"mock-client-id\" got [\"different-client-id\"]") {
				t.Fatalf("error is nil or error message does not match expected error")
			}

			if tt.name == "Success" && (err != nil || token.Subject != "mock-subject") {
				t.Fatalf("expected token does not match returned token")
			}
		})
	}
}

func TestJWKSTokenVerifier_VerifyIDToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.JWKSTokenVerifier.VerifyIDToken")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	mockToken := new(oidc.IDToken)
	mockToken.Subject = "mock-subject"

	for _, tt := range []struct {
		name  string
		token string
	}{
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJtb2NrLWNsaWVudC1pZCIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.BdspASNsnxeXnqZXZnFnkvv-ClMq0U6X1gCIUrh9V7c",
		},
		{
			name:  "Fail",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJkaWZmZXJlbnQtY2xpZW50LWlkIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.ZC8qlDOhxOZyUTiAk2ICpEmSnmEb2UHXAJyOrrYVCDc",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockProviderInterface(ctrl)
			tokenVerifier := oidc.NewVerifier("", nil, &oidc.Config{
				ClientID:                   "mock-client-id",
				SkipExpiryCheck:            true,
				SkipIssuerCheck:            true,
				InsecureSkipSignatureCheck: true,
			})
			mockProvider.EXPECT().Verifier(&oidc.Config{ClientID: "mock-client-id"}).Return(tokenVerifier)

			verifier := NewJWKSTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor)

			token, err := verifier.VerifyIDToken(context.TODO(), tt.token)

			if tt.name == "Fail" && (err == nil || err.Error() != "oidc: expected audience \"mock-client-id\" got [\"different-client-id\"]") {
				t.Fatalf("error is nil or error message does not match expected error")
			}

			if tt.name == "Success" && (err != nil || token.Subject != "mock-subject") {
				t.Fatalf("expected token does not match returned token")
			}
		})
	}
}

func TestUserinfoTokenVerifier_VerifyAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.UserinfoTokenVerifier.VerifyAccessToken")).
		Times(1).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	for _, tt := range []struct {
		name     string
		token    string
		userinfo *oidc.UserInfo
		err      error
	}{
		{
			name:     "Fail",
			token:    "mock-opaque-access-token",
			userinfo: nil,
			err:      errors.New("mock-error"),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockProviderInterface(ctrl)
			tokenVerifier := oidc.NewVerifier("", nil, &oidc.Config{
				ClientID:                   "mock-client-id",
				SkipExpiryCheck:            true,
				SkipIssuerCheck:            true,
				InsecureSkipSignatureCheck: true,
			})
			mockProvider.EXPECT().Verifier(&oidc.Config{ClientID: "mock-client-id"}).Return(tokenVerifier)
			mockProvider.EXPECT().UserInfo(gomock.Any(), gomock.Any()).Return(tt.userinfo, tt.err)

			verifier := NewUserinfoTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor)

			_, err := verifier.VerifyAccessToken(context.TODO(), tt.token)

			if tt.name == "Fail" && (err == nil || err.Error() != "mock-error") {
				t.Fatalf("error is nil or error message does not match expected error")
			}
		})

	}
}

func TestUserinfoTokenVerifier_VerifyIDToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockTracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("authentication.UserinfoTokenVerifier.VerifyIDToken")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockMonitor := NewMockMonitorInterface(ctrl)

	mockToken := new(oidc.IDToken)
	mockToken.Subject = "mock-subject"

	for _, tt := range []struct {
		name  string
		token string
	}{
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJtb2NrLWNsaWVudC1pZCIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.BdspASNsnxeXnqZXZnFnkvv-ClMq0U6X1gCIUrh9V7c",
		},
		{
			name:  "Fail",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJkaWZmZXJlbnQtY2xpZW50LWlkIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.ZC8qlDOhxOZyUTiAk2ICpEmSnmEb2UHXAJyOrrYVCDc",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockProviderInterface(ctrl)
			tokenVerifier := oidc.NewVerifier("", nil, &oidc.Config{
				ClientID:                   "mock-client-id",
				SkipExpiryCheck:            true,
				SkipIssuerCheck:            true,
				InsecureSkipSignatureCheck: true,
			})
			mockProvider.EXPECT().Verifier(&oidc.Config{ClientID: "mock-client-id"}).Return(tokenVerifier)

			verifier := NewUserinfoTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor)

			token, err := verifier.VerifyIDToken(context.TODO(), tt.token)

			if tt.name == "Fail" && (err == nil || err.Error() != "oidc: expected audience \"mock-client-id\" got [\"different-client-id\"]") {
				t.Fatalf("error is nil or error message does not match expected error")
			}

			if tt.name == "Success" && (err != nil || token.Subject != "mock-subject") {
				t.Fatalf("expected token does not match returned token")
			}
		})
	}
}
