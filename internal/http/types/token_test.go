// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package types -destination ./mock_logger.go -source=../../logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package types -destination ./mock_tracer.go -source=../../tracing/interfaces.go

func TestLoadFromRequestSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTracer := NewMockTracingInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)

	p := TokenPaginator{
		tokens: make(map[string]string),
		tracer: mockTracer,
		logger: mockLogger,
	}

	mockTracer.EXPECT().Start(gomock.Eq(context.TODO()), gomock.Any()).Return(nil, trace.SpanFromContext(context.TODO()))

	mockRequest := httptest.NewRequest(http.MethodGet, "/path/to/endpoint", nil)
	// base64("{"token-1": "continuation-token-1","token-2":"continuation-token-2","token-3":""}")
	const mockValue = "eyJ0b2tlbi0xIjogImNvbnRpbnVhdGlvbi10b2tlbi0xIiwidG9rZW4tMiI6ImNvbnRpbnVhdGlvbi10b2tlbi0yIiwidG9rZW4tMyI6IiJ9"
	mockRequest.Header.Set(
		"X-Token-Pagination",
		mockValue,
	)

	err := p.LoadFromRequest(context.TODO(), mockRequest)
	if err != nil {
		t.Errorf("Unexpected error while running LoadFromRequest")
	}

	expectedMap := make(map[string]string, 2)
	expectedMap["token-1"] = "continuation-token-1"
	expectedMap["token-2"] = "continuation-token-2"

	if !reflect.DeepEqual(p.tokens, expectedMap) {
		t.Fail()
	}
}

func TestLoadFromRequestSuccessNoHeader(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTracer := NewMockTracingInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)

	p := TokenPaginator{
		tokens: make(map[string]string),
		tracer: mockTracer,
		logger: mockLogger,
	}

	mockTracer.EXPECT().Start(gomock.Eq(context.TODO()), gomock.Any()).Return(nil, trace.SpanFromContext(context.TODO()))

	mockRequest := httptest.NewRequest(http.MethodGet, "/path/to/endpoint", nil)

	err := p.LoadFromRequest(context.TODO(), mockRequest)
	if err != nil {
		t.Errorf("Unexpected error while running LoadFromRequest")
	}

	if p.tokens == nil {
		t.Fail()
	}
}

func TestLoadFromRequestFailure_WrongHeaderValue(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTracer := NewMockTracingInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)

	p := TokenPaginator{
		tokens: make(map[string]string),
		tracer: mockTracer,
		logger: mockLogger,
	}

	mockTracer.EXPECT().Start(gomock.Eq(context.TODO()), gomock.Any()).Return(nil, trace.SpanFromContext(context.TODO()))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())

	mockRequest := httptest.NewRequest(http.MethodGet, "/path/to/endpoint", nil)
	// base64("definitely not a json string")
	const mockValue = "ZGVmaW5pdGVseSBub3QgYSBqc29uIHN0cmluZw=="
	mockRequest.Header.Set(
		"X-Token-Pagination",
		mockValue,
	)

	err := p.LoadFromRequest(context.TODO(), mockRequest)
	if err == nil {
		t.Fail()
	}
}
