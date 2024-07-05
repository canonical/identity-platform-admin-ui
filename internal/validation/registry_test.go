// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package validation -destination ./mock_logger.go -source=../logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package validation -destination ./mock_monitor.go -source=../monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package validation -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

type payloadValidator struct{}

func (p *payloadValidator) Validate(ctx context.Context, _, _ string, _ []byte) (context.Context, validator.ValidationErrors, error) {
	e := mockValidationErrors()
	if e == nil {
		return ctx, nil, nil
	}
	return ctx, e, nil
}

func (p *payloadValidator) NeedsValidation(r *http.Request) bool {
	return true
}

type noopPayloadValidator struct{}

func (_ *noopPayloadValidator) Validate(ctx context.Context, _, _ string, _ []byte) (context.Context, validator.ValidationErrors, error) {
	return ctx, nil, nil
}

func (_ *noopPayloadValidator) NeedsValidation(r *http.Request) bool {
	return true
}

func TestValidator_Middleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)
	logger := NewMockLoggerInterface(ctrl)

	tracer.EXPECT().
		Start(gomock.Any(), gomock.Eq("validator.ValidationRegistry.ValidationMiddleware")).
		Times(2).
		Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("main handler\n"))
	})

	vld := NewRegistry(tracer, monitor, logger)
	vld.validators["mock-key"] = &payloadValidator{}

	for _, tt := range []struct {
		name            string
		expctedResponse string
		expectedCode    int
		mockRequest     *http.Request
	}{
		{
			name:         "MatchingPrefix",
			expectedCode: http.StatusBadRequest,
			mockRequest:  httptest.NewRequest(http.MethodGet, "/api/v0/mock-key", nil),
		},
		{
			name:            "NoMatchingPrefix",
			expctedResponse: "main handler\n",
			expectedCode:    http.StatusOK,
			mockRequest:     httptest.NewRequest(http.MethodGet, "/api/v0/different-mock-key", nil),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockResponse := httptest.NewRecorder()

			decoratedHandler := vld.ValidationMiddleware(mainHandler)
			decoratedHandler.ServeHTTP(mockResponse, tt.mockRequest)

			responseValue := mockResponse.Body.String()

			if mockResponse.Code != tt.expectedCode {
				t.Fatalf("actual response code differes from expected")
			}

			if tt.expctedResponse == "" {
				return
			}

			if responseValue != tt.expctedResponse {
				t.Fatalf("actual response body differes from expected")
			}
		})
	}
}

func mockValidationErrors() validator.ValidationErrors {
	type InvalidStruct struct {
		FirstName string `json:"first_name" validate:"required"`
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	e := validate.Struct(InvalidStruct{})
	if e == nil {
		return nil
	}
	return e.(validator.ValidationErrors)
}

func TestValidator_RegisterValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)
	logger := NewMockLoggerInterface(ctrl)

	emptyValidator := &ValidationRegistry{
		validators: make(map[string]PayloadValidatorInterface),
		tracer:     tracer,
		monitor:    monitor,
		logger:     logger,
	}

	noopValidator := &noopPayloadValidator{}

	validators := make(map[string]PayloadValidatorInterface)
	validators["mock-key-1"] = noopValidator

	nonEmptyValidator := &ValidationRegistry{
		validators: validators,
		tracer:     tracer,
		monitor:    monitor,
		logger:     logger,
	}

	for _, tt := range []struct {
		name      string
		validator *ValidationRegistry
		prefix    string
		v         PayloadValidatorInterface
		expected  string
	}{
		{
			name:      "Nil middleware",
			validator: emptyValidator,
			prefix:    "",
			v:         nil,
			expected:  "payloadValidator can't be null",
		},
		{
			name:      "Existing key",
			validator: nonEmptyValidator,
			prefix:    "mock-key-1",
			v:         noopValidator,
			expected:  "key is already registered",
		},
		{
			name:      "Success",
			validator: emptyValidator,
			prefix:    "mock-key",
			v:         noopValidator,
			expected:  "",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validator.RegisterPayloadValidator(tt.prefix, tt.v)

			if tt.expected == "" && nil == result {
				return
			}

			if result.Error() != tt.expected {
				t.Fatalf("returned error doesn't match expected error")
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)
	logger := NewMockLoggerInterface(ctrl)

	v := NewRegistry(tracer, monitor, logger)

	if v.tracer != tracer {
		t.FailNow()
	}

	if v.monitor != monitor {
		t.FailNow()
	}

	if v.logger != logger {
		t.FailNow()
	}

	if v.validators == nil {
		t.Fatalf("validators map expected not empty")
	}

	if len(v.validators) != 0 {
		t.Fatalf("validators map expected not populated")
	}
}

func TestNewValidationError(t *testing.T) {
	ve := mockValidationErrors()
	response := NewValidationError("validation errors", ve)

	if response.Status != http.StatusBadRequest {
		t.Fatalf("response status does not match expected")
	}

	if response.Message != "validation errors" {
		t.Fatalf("response message does not match expected")
	}

	expectedData := map[string][]string{"first_name": {"Missing required field 'first_name'"}}

	if !reflect.DeepEqual(expectedData, response.Data) {
		t.Fatalf("Expected '%s', got '%s'", expectedData, response.Data)
	}
}
