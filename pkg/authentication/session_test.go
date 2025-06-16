// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	reflect "reflect"
	"testing"

	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_kratos.go github.com/ory/kratos-client-go IdentityAPI
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_kratos_frontend_api.go github.com/ory/kratos-client-go FrontendAPI

func TestGetIdentitySessionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockKratosFrontendAPI := NewMockFrontendAPI(ctrl)

	ctx := context.Background()
	cookies := make([]*http.Cookie, 0)
	cookie := &http.Cookie{Name: "test", Value: "test"}
	cookies = append(cookies, cookie)
	session := &kClient.Session{
		Id: "test",
	}

	mockTracer.EXPECT().Start(ctx, "authentication.SessionManager.GetIdentitySession").Return(ctx, trace.SpanFromContext(ctx))

	sessionRequest := kClient.FrontendAPIToSessionRequest{
		ApiService: mockKratosFrontendAPI,
	}
	resp := http.Response{
		Header: http.Header{"Set-Cookie": []string{cookie.Raw}},
	}

	mockKratosFrontendAPI.EXPECT().ToSession(ctx).Times(1).Return(sessionRequest)
	mockKratosFrontendAPI.EXPECT().ToSessionExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.FrontendAPIToSessionRequest) (*kClient.Session, *http.Response, error) {
			if cookie := (*string)(reflect.ValueOf(r).FieldByName("cookie").UnsafePointer()); *cookie != "test=test" {
				t.Fatalf("expected cookie string as test=test, got %s", *cookie)
			}

			return session, &resp, nil
		},
	)

	service := NewSessionManagerService(mockKratosIdentityAPI, mockKratosFrontendAPI, mockTracer, mockMonitor, mockLogger)

	s, err := service.GetIdentitySession(ctx, cookies)

	if err != nil {
		t.Fatalf("expected error to be nil, got %v", err)
	}

	if s.Error != nil {
		t.Fatalf("expected session error to be nil, got %v", s.Error)
	}

	if s.Session.Id != session.Id {
		t.Fatalf("expected session id %v, got %v", session.Id, s.Session.Id)
	}
}

func TestGetIdentitySessionFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockKratosFrontendAPI := NewMockFrontendAPI(ctrl)

	ctx := context.Background()
	cookies := make([]*http.Cookie, 0)
	cookie := &http.Cookie{Name: "test", Value: "test"}
	cookies = append(cookies, cookie)

	mockTracer.EXPECT().Start(ctx, "authentication.SessionManager.GetIdentitySession").Return(ctx, trace.SpanFromContext(ctx))

	sessionRequest := kClient.FrontendAPIToSessionRequest{
		ApiService: mockKratosFrontendAPI,
	}

	rr := httptest.NewRecorder()
	rr.Header().Set("Content-Type", "application/json")
	rr.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(rr).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusUnauthorized,
			"debug":   "invalid kratos session cookie",
			"details": map[string]interface{}{},
			"id":      "some-id",
			"message": "unauthorized",
			"reason":  "invalid credentials",
			"request": "req-id-123",
			"status":  "Unauthorized",
		},
	})
	resp := rr.Result()
	resp.Body = io.NopCloser(bytes.NewBuffer(rr.Body.Bytes()))

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	mockKratosFrontendAPI.EXPECT().ToSession(ctx).Return(sessionRequest).Times(1)
	mockKratosFrontendAPI.EXPECT().
		ToSessionExecute(gomock.Any()).
		Times(1).
		DoAndReturn(func(r kClient.FrontendAPIToSessionRequest) (*kClient.Session, *http.Response, error) {
			return nil, resp, fmt.Errorf("error")
		})

	service := NewSessionManagerService(mockKratosIdentityAPI, mockKratosFrontendAPI, mockTracer, mockMonitor, mockLogger)

	s, err := service.GetIdentitySession(ctx, cookies)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if s.Error == nil {
		t.Fatal("expected session error to be populated")
	}

	if *s.Error.Code != int64(http.StatusUnauthorized) {
		t.Fatalf("expected error code %v, got %v", http.StatusUnauthorized, *s.Error.Code)
	}
}

func TestDisableSessionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockKratosFrontendAPI := NewMockFrontendAPI(ctrl)

	ctx := context.Background()
	sessionID := "test"

	mockTracer.EXPECT().Start(ctx, "authentication.SessionManager.DisableSession").Return(ctx, trace.SpanFromContext(ctx))

	disableSessionRequest := kClient.IdentityAPIDisableSessionRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockKratosIdentityAPI.EXPECT().DisableSession(ctx, sessionID).Return(disableSessionRequest)
	mockKratosIdentityAPI.EXPECT().DisableSessionExecute(disableSessionRequest).Return(&http.Response{StatusCode: http.StatusNoContent}, nil)

	service := NewSessionManagerService(mockKratosIdentityAPI, mockKratosFrontendAPI, mockTracer, mockMonitor, mockLogger)

	session, err := service.DisableSession(ctx, sessionID)

	if err != nil {
		t.Fatalf("expected error to be nil, got: %v", err)
	}
	if session == nil {
		t.Fatalf("expected session data, got nil")
	}
}

func TestDisableSessionFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockKratosFrontendAPI := NewMockFrontendAPI(ctrl)

	ctx := context.Background()
	sessionID := "test"

	disableSessionRequest := kClient.IdentityAPIDisableSessionRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockTracer.EXPECT().Start(ctx, "authentication.SessionManager.DisableSession").AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().DisableSession(ctx, sessionID).Times(1).Return(disableSessionRequest)

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	mockKratosIdentityAPI.EXPECT().
		DisableSessionExecute(gomock.Any()).
		Times(1).
		DoAndReturn(func(r kClient.IdentityAPIDisableSessionRequest) (*http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusBadRequest,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "some-id",
						"message": "session not found",
						"reason":  "not found",
						"request": "req-id-123",
						"status":  "Not Found",
					},
				},
			)

			res := rr.Result()

			bodyBytes := rr.Body.Bytes()
			res.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			return res, fmt.Errorf("error")
		})

	service := NewSessionManagerService(mockKratosIdentityAPI, mockKratosFrontendAPI, mockTracer, mockMonitor, mockLogger)

	s, err := service.DisableSession(ctx, sessionID)

	if err == nil {
		t.Fatalf("expected error, not nil")
	}

	if s.Error == nil {
		t.Fatalf("expected session error to be populated")
	}

	if *s.Error.Code != int64(http.StatusBadRequest) {
		t.Fatalf("expected error code %v, got %v", http.StatusBadRequest, *s.Error.Code)
	}
}
