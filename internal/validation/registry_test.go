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
	vld.validatingFuncs["mock-key"] = func(r *http.Request) (validator.ValidationErrors, error) {
		e := mockValidationErrors()
		if e == nil {
			return nil, nil
		}
		return e, nil
	}

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
		validatingFuncs: make(map[string]ValidatingFunc),
		tracer:          tracer,
		monitor:         monitor,
		logger:          logger,
	}

	noopVf := ValidatingFunc(func(r *http.Request) (validator.ValidationErrors, error) {
		return nil, nil
	})
	validatingFuncs := make(map[string]ValidatingFunc)
	validatingFuncs["mock-key-1"] = noopVf

	nonEmptyValidator := &ValidationRegistry{
		validatingFuncs: validatingFuncs,
		tracer:          tracer,
		monitor:         monitor,
		logger:          logger,
	}

	for _, tt := range []struct {
		name      string
		validator *ValidationRegistry
		prefix    string
		vf        ValidatingFunc
		expected  string
	}{
		{
			name:      "Nil middleware",
			validator: emptyValidator,
			prefix:    "",
			vf:        nil,
			expected:  "validatingFunc can't be null",
		},
		{
			name:      "Existing key",
			validator: nonEmptyValidator,
			prefix:    "mock-key-1",
			vf:        noopVf,
			expected:  "key is already registered",
		},
		{
			name:      "Success",
			validator: emptyValidator,
			prefix:    "mock-key",
			vf:        noopVf,
			expected:  "",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validator.RegisterValidatingFunc(tt.prefix, tt.vf)

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

	if v.validatingFuncs == nil {
		t.Fatalf("validatingFuncs map expected not empty")
	}

	if len(v.validatingFuncs) != 0 {
		t.Fatalf("validatingFuncs map expected not populated")
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

	expectedData := map[string][]string{
		"first_name": {
			"value '' fails validation of type `required`",
		},
	}

	if !reflect.DeepEqual(expectedData, response.Data) {
		t.Fatalf("response data does not match expected validation errors")
	}
}
