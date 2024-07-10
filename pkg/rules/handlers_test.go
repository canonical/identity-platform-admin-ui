// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	oathkeeper "github.com/ory/oathkeeper-client-go"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
)

//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_corev1.go k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,ConfigMapInterface
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_oathkeeper.go github.com/ory/oathkeeper-client-go ApiApi
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_validation.go -source=../../internal/validation/registry.go

func TestHandleListSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_id2 := "mocked_rule2:deny"
	string_authorizer_handler_deny := "deny"
	string_authorizer_handler_allow := "allow"

	serviceOutput := []oathkeeper.Rule{
		{
			Id: &string_id1,
			Authenticators: []oathkeeper.RuleHandler{
				{
					Handler: &string_athenticator_handler,
				},
			},
			Mutators: []oathkeeper.RuleHandler{
				{
					Handler: &string_mutator_handler,
				},
			},
			Authorizer: &oathkeeper.RuleHandler{
				Handler: &string_authorizer_handler_allow,
			},
			Match: &oathkeeper.RuleMatch{
				Methods: []string{"GET", "POST"},
			},
		},
		{
			Id: &string_id2,
			Authenticators: []oathkeeper.RuleHandler{
				{
					Handler: &string_athenticator_handler,
				},
			},
			Mutators: []oathkeeper.RuleHandler{
				{
					Handler: &string_mutator_handler,
				},
			},
			Authorizer: &oathkeeper.RuleHandler{
				Handler: &string_authorizer_handler_deny,
			},
			Match: &oathkeeper.RuleMatch{
				Methods: []string{"PUT", "DELETE"},
			},
		},
	}

	var offset int64 = 0
	var size int64 = 100

	mockService.EXPECT().ListRules(gomock.Any(), offset, size).Return(serviceOutput, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/rules?page_token=eyJvZmZzZXQiOjB9&size=100", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusOK, res.StatusCode)
	}

	type Response struct {
		Data []oathkeeper.Rule `json:"data"`
	}

	rr := new(Response)

	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	for i, r := range rr.Data {
		if *r.Id != *serviceOutput[i].Id {
			t.Fatalf("invalid result, expected: %v, got: %v", *serviceOutput[i].Id, *r.Id)
		}
	}
}

func TestHandleListFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	var offset int64 = 0
	var size int64 = 100

	mockService.EXPECT().ListRules(gomock.Any(), offset, size).Return(nil, fmt.Errorf("mock_error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/rules?pageToken=0&offset=100", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, res.StatusCode)
	}

	rr := new(types.Response)
	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Status != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}

	if rr.Message != "mock_error" {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}
}

func TestHandleDetailSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_authorizer_handler_allow := "allow"

	serviceOutput := []oathkeeper.Rule{
		{
			Id: &string_id1,
			Authenticators: []oathkeeper.RuleHandler{
				{
					Handler: &string_athenticator_handler,
				},
			},
			Mutators: []oathkeeper.RuleHandler{
				{
					Handler: &string_mutator_handler,
				},
			},
			Authorizer: &oathkeeper.RuleHandler{
				Handler: &string_authorizer_handler_allow,
			},
			Match: &oathkeeper.RuleMatch{
				Methods: []string{"GET", "POST"},
			},
		},
	}

	mockService.EXPECT().GetRule(gomock.Any(), gomock.Any()).Return(serviceOutput, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/rules/mocked_rule1:allow", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusOK, res.StatusCode)
	}

	type Response struct {
		Data []oathkeeper.Rule `json:"data"`
	}

	rr := new(Response)

	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	for i, r := range rr.Data {
		if *r.Id != *serviceOutput[i].Id {
			t.Fatalf("invalid result, expected: %v, got: %v", *serviceOutput[i].Id, *r.Id)
		}
	}
}

func TestHandleDetailFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	mockService.EXPECT().GetRule(gomock.Any(), "mocked_rule1:allow").Return(nil, fmt.Errorf("mock_error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/rules/mocked_rule1:allow", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, res.StatusCode)
	}

	rr := new(types.Response)
	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Status != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}

	if rr.Message != "mock_error" {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}
}

func TestHandleUpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_authorizer_handler_allow := "allow"

	ruleUpdate := oathkeeper.Rule{
		Id: &string_id1,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &string_athenticator_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &string_mutator_handler,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &string_authorizer_handler_allow,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	payload, _ := json.Marshal(ruleUpdate)

	mockService.EXPECT().UpdateRule(gomock.Any(), "mocked_rule1:allow", gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v0/rules/mocked_rule1:allow", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusOK, res.StatusCode)
	}

	rr := new(types.Response)

	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Message != fmt.Sprintf("Updated rule %s", *ruleUpdate.Id) {
		t.Fatalf("invalid result, expected: %v, got: %v", ruleUpdate.Id, rr.Message)
	}
}

func TestHandleUpdateFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_authorizer_handler_allow := "allow"

	ruleUpdate := oathkeeper.Rule{
		Id: &string_id1,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &string_athenticator_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &string_mutator_handler,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &string_authorizer_handler_allow,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	payload, _ := json.Marshal(ruleUpdate)

	mockService.EXPECT().UpdateRule(gomock.Any(), "mocked_rule1:allow", gomock.Any()).Return(fmt.Errorf("mock_error"))

	req := httptest.NewRequest(http.MethodPut, "/api/v0/rules/mocked_rule1:allow", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, res.StatusCode)
	}

	rr := new(types.Response)
	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Status != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}

	if rr.Message != "mock_error" {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}
}

func TestHandleCreateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_authorizer_handler_allow := "allow"

	ruleCreated := oathkeeper.Rule{
		Id: &string_id1,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &string_athenticator_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &string_mutator_handler,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &string_authorizer_handler_allow,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	payload, _ := json.Marshal(ruleCreated)

	mockService.EXPECT().CreateRule(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v0/rules", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusCreated, res.StatusCode)
	}

	rr := new(types.Response)

	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Message != fmt.Sprintf("Created rule %s", *ruleCreated.Id) {
		t.Fatalf("invalid result, expected: %v, got: %v", ruleCreated.Id, rr.Message)
	}
}

func TestHandleCreateFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id1 := "mocked_rule1:allow"
	string_authorizer_handler_allow := "allow"

	ruleCreated := oathkeeper.Rule{
		Id: &string_id1,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &string_athenticator_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &string_mutator_handler,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &string_authorizer_handler_allow,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	payload, _ := json.Marshal(ruleCreated)

	mockService.EXPECT().CreateRule(gomock.Any(), gomock.Any()).Return(fmt.Errorf("mock_error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v0/rules", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, res.StatusCode)
	}

	rr := new(types.Response)
	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Status != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}

	if rr.Message != "mock_error" {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}
}

func TestHandleRemoveSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	mockService.EXPECT().DeleteRule(gomock.Any(), "mocked_rule1:allow").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v0/rules/mocked_rule1:allow", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusOK, res.StatusCode)
	}

	rr := new(types.Response)

	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Message != fmt.Sprintf("Deleted rule mocked_rule1:allow") {
		t.Fatalf("invalid result, expected: Deleted rule mocked_rule1:allow, got: %v", rr.Message)
	}
}

func TestHandleRemoveFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)

	mockService.EXPECT().DeleteRule(gomock.Any(), "mocked_rule1:allow").Return(fmt.Errorf("mock_error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v0/rules/mocked_rule1:allow", nil)
	w := httptest.NewRecorder()
	mux := chi.NewMux()
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, res.StatusCode)
	}

	rr := new(types.Response)
	if err := json.Unmarshal(data, rr); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if rr.Status != http.StatusInternalServerError {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}

	if rr.Message != "mock_error" {
		t.Fatalf("expected HTTP status code %v got %v", http.StatusInternalServerError, rr.Status)
	}
}

func TestRegisterValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)
	mockValidationRegistry := NewMockValidationRegistryInterface(ctrl)

	apiKey := "rules"
	mockValidationRegistry.EXPECT().
		RegisterPayloadValidator(gomock.Eq(apiKey), gomock.Any()).
		Return(nil)
	mockValidationRegistry.EXPECT().
		RegisterPayloadValidator(gomock.Eq(apiKey), gomock.Any()).
		Return(fmt.Errorf("key is already registered"))

	// first registration of `apiKey` is successful
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterValidation(mockValidationRegistry)

	mockLogger.EXPECT().Fatalf(gomock.Any(), gomock.Any()).Times(1)

	// second registration of `apiKey` causes logger.Fatal invocation
	NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterValidation(mockValidationRegistry)
}
