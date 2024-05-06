// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_validation.go -source=../../internal/validation/registry.go

// + http :8000/api/v0/roles X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 97
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:32 GMT

// {
//     "_meta": null,
//     "data": [
//         "global",
//         "administrator",
//         "viewer"
//     ],
//     "message": "List of roles",
//     "status": 200
// }

func TestHandleList(t *testing.T) {
	type expected struct {
		err   error
		roles []string
	}

	tests := []struct {
		name     string
		expected expected
		output   *types.Response
	}{
		{
			name: "empty result",
			expected: expected{
				roles: []string{},
				err:   nil,
			},
			output: &types.Response{
				Data:    []string{},
				Message: "List of roles",
				Status:  http.StatusOK,
			},
		},
		{
			name: "error",
			expected: expected{
				roles: []string{},
				err:   fmt.Errorf("error"),
			},
			output: &types.Response{
				Data:    []string{},
				Message: "error",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name: "full result",
			expected: expected{
				roles: []string{"global", "administrator", "viewer"},
				err:   nil,
			},

			output: &types.Response{
				Data:    []string{"global", "administrator", "viewer"},
				Message: "List of roles",
				Status:  http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodGet, "/api/v0/roles", nil)

			mockService.EXPECT().ListRoles(gomock.Any(), "anonymous").Return(test.expected.roles, test.expected.err)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(rr.Data, test.output.Data) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http :8000/api/v0/roles/administrator X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 77
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:32 GMT

// {
//     "_meta": null,
//     "data": [
//         "administrator"
//     ],
//     "message": "Rule detail",
//     "status": 200
// }

func TestHandleDetail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "unknown role",
			input:    "unknown",
			expected: fmt.Errorf("role does not exist"),
			output: &types.Response{
				Message: "role does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Data: []Role{{
					ID:   "administrator",
					Name: "administrator",
				}},
				Message: "Rule detail",
				Status:  http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s", test.input), nil)

			mockService.EXPECT().GetRole(gomock.Any(), "anonymous", test.input).DoAndReturn(
				func(context.Context, string, string) (*Role, error) {
					if test.expected != nil {
						return nil, test.expected
					}

					return &Role{
						ID:   test.input,
						Name: test.input,
					}, nil

				},
			)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []Role            `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && !reflect.DeepEqual(rr.Data, test.output.Data) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestHandleUpdate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "unknown role",
			input:    "unknown",
			expected: fmt.Errorf("role does not exist"),
			output: &types.Response{
				Message: "use /api/v0/roles/unknown/entitlements to assign permissions",
				Status:  http.StatusNotImplemented,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Message: "use /api/v0/roles/administrator/entitlements to assign permissions",
				Status:  http.StatusNotImplemented,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/roles/%s", test.input), nil)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Message string `json:"message"`
				Status  int    `json:"status"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http :8000/api/v0/roles/administrator/entitlements X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 156
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

// {
//     "_meta": null,
//     "data": [
//        "can_view::client:github-canonical",
//        "can_delete::client:okta",
//        "can_edit::client:okta"
//     ],
//     "message": "List of entitlements",
//     "status": 200
// }

func TestHandleListPermissionsSuccess(t *testing.T) {
	type expected struct {
		permissions []string
		cTokens     map[string]string
	}

	tests := []struct {
		name     string
		expected expected
		output   *types.Response
	}{
		{
			name:     "no permissions",
			expected: expected{permissions: []string{}},
			output: &types.Response{
				Data:    []string{},
				Message: "List of entitlements",
				Status:  http.StatusOK,
			},
		},
		{
			name: "full permissions",
			expected: expected{
				permissions: []string{
					"can_view::client:github-canonical",
					"can_delete::client:okta",
					"can_edit::client:okta",
				},
				cTokens: map[string]string{"client": "test"},
			},
			output: &types.Response{
				Data: []string{
					"can_view::client:github-canonical",
					"can_delete::client:okta",
					"can_edit::client:okta",
				},
				Message: "List of entitlements",
				Status:  http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			roleID := "administrator"
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s/entitlements", roleID), nil)

			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockService.EXPECT().ListPermissions(gomock.Any(), roleID, map[string]string{}).Return(test.expected.permissions, test.expected.cTokens, nil)

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
				t.Errorf("expected HTTP status code 200 got %v", res.StatusCode)
			}

			tokenMap, err := base64.StdEncoding.DecodeString(res.Header.Get(types.PAGINATION_HEADER))

			if test.expected.cTokens != nil {
				if err != nil {
					t.Errorf("expected continuation token in headers")
				}

				tokens := map[string]string{}

				_ = json.Unmarshal(tokenMap, &tokens)

				if !reflect.DeepEqual(tokens, test.expected.cTokens) {
					t.Errorf("expected continuation tokens to match: %v - %v", tokens, test.expected.cTokens)
				}
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if !reflect.DeepEqual(rr.Data, test.output.Data) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http :8000/api/v0/roles/administrator/groups X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 87
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:35 GMT

// {
//     "_meta": null,
//     "data": [
//         "group:c-level#member"
//     ],
//     "message": "List of groups",
//     "status": 200
// }

func TestHandleListRoleGroupsSuccess(t *testing.T) {
	type expected struct {
		groups  []string
		cTokens map[string]string
	}

	tests := []struct {
		name     string
		expected expected
		output   *types.Response
	}{
		{
			name:     "no groups",
			expected: expected{groups: []string{}},
			output: &types.Response{
				Data:    []string{},
				Message: "List of groups",
				Status:  http.StatusOK,
			},
		},
		{
			name: "full groups",
			expected: expected{
				groups: []string{
					"group:c-level#member",
					"group:it-admin#member",
				},
				cTokens: map[string]string{"roles": "test"},
			},
			output: &types.Response{
				Data: []string{
					"group:c-level#member",
					"group:it-admin#member",
				},
				Message: "List of groups",
				Status:  http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			roleID := "administrator"
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s/groups", roleID), nil)

			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockService.EXPECT().ListRoleGroups(gomock.Any(), roleID, "").Return(test.expected.groups, test.expected.cTokens["roles"], nil)

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
				t.Errorf("expected HTTP status code 200 got %v", res.StatusCode)
			}

			tokenMap, err := base64.StdEncoding.DecodeString(res.Header.Get(types.PAGINATION_HEADER))

			if test.expected.cTokens != nil {
				if err != nil {
					t.Errorf("expected continuation token in headers")
				}

				tokens := map[string]string{}

				_ = json.Unmarshal(tokenMap, &tokens)

				if !reflect.DeepEqual(tokens, test.expected.cTokens) {
					t.Errorf("expected continuation tokens to match: %v - %v", tokens, test.expected.cTokens)
				}
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if !reflect.DeepEqual(rr.Data, test.output.Data) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http DELETE :8000/api/v0/roles/administrator/entitlements/can_edit::client:okta X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 116
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Removed permission can_edit::client:okta for role administrator",
//	    "status": 200
//	}

func TestHandleRemovePermissionBadPermissionFormat(t *testing.T) {
	type input struct {
		roleID       string
		permissionID string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name: "wrong permission format",
			input: input{
				roleID:       "administrator",
				permissionID: "can_edit-something-wrong:okta",
			},
			expected: fmt.Errorf("role does not exist"),
			output: &types.Response{
				Message: "Error parsing entitlement ID",
				Status:  http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s/entitlements/%s", test.input.roleID, test.input.permissionID), nil)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && !reflect.DeepEqual(rr.Data, test.output.Data) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestHandleRemovePermission(t *testing.T) {
	type input struct {
		roleID       string
		permissionID string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name: "unknown role",
			input: input{
				roleID:       "unknown",
				permissionID: "can_edit::client::okta",
			},
			expected: fmt.Errorf("role does not exist"),
			output: &types.Response{
				Message: "role does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name: "found",
			input: input{
				roleID:       "administrator",
				permissionID: "can_edit::client:okta",
			},
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Removed permission can_edit::client:okta for role administrator",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s/entitlements/%s", test.input.roleID, test.input.permissionID), nil)

			mockService.EXPECT().RemovePermissions(
				gomock.Any(),
				test.input.roleID,
				Permission{
					Relation: strings.Split(test.input.permissionID, authorization.PERMISSION_SEPARATOR)[0],
					Object:   strings.Split(test.input.permissionID, authorization.PERMISSION_SEPARATOR)[1],
				},
			).Return(test.expected)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && len(rr.Data) != 0 {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http PATCH :8000/api/v0/roles/administrator/entitlements 'permissions:=[{"relation":"can_delete","object":"scheme:superman"},{"relation":"can_view","object":"client:aws"}]' X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 201 Created
// Content-Length: 95
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:34 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Updated permissions for role administrator",
//	    "status": 201
//	}
func TestHandleAssignPermissions(t *testing.T) {
	type input struct {
		permissions []Permission
		roleID      string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name:     "multiple permissions",
			expected: nil,
			input: input{
				roleID: "administrator",
				permissions: []Permission{
					{
						Relation: "can_view",
						Object:   "client:github-canonical",
					},
					{
						Relation: "can_delete",
						Object:   "client:okta",
					},
					{
						Relation: "can_edit",
						Object:   "client:okta",
					},
				},
			},
			output: &types.Response{
				Message: "Updated permissions for role administrator",
				Status:  http.StatusCreated,
			},
		},
		{
			name:     "multiple permissions with error",
			expected: fmt.Errorf("error"),
			input: input{
				roleID: "administrator",
				permissions: []Permission{
					{
						Relation: "can_view",
						Object:   "client:github-canonical",
					},
					{
						Relation: "can_delete",
						Object:   "client:okta",
					},
					{
						Relation: "can_edit",
						Object:   "client:okta",
					},
				},
			},
			output: &types.Response{
				Message: "error",
				Status:  http.StatusInternalServerError,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			upr := new(UpdatePermissionsRequest)
			upr.Permissions = test.input.permissions
			payload, _ := json.Marshal(upr)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/roles/%s/entitlements", test.input.roleID), bytes.NewReader(payload))

			mockService.EXPECT().AssignPermissions(gomock.Any(), test.input.roleID, test.input.permissions).Return(test.expected)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && len(rr.Data) != 0 {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestHandleAssignPermissionsBadPermissionFormat(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "no permissions",
			expected: nil,
			input:    "administrator",
			output: &types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/roles/%s/entitlements", test.input), nil)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

// + http DELETE :8000/api/v0/roles/viewer X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 72
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:36 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Deleted role viewer",
//	    "status": 200
//	}
func TestHandleRemove(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "unknown role",
			input:    "unknown",
			expected: fmt.Errorf("role does not exist"),
			output: &types.Response{
				Message: "role does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Deleted role administrator",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s", test.input), nil)

			mockService.EXPECT().DeleteRole(
				gomock.Any(),
				test.input,
			).Return(test.expected)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && len(rr.Data) != 0 {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestHandleCreate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "success",
			expected: nil,
			input:    "administrator",

			output: &types.Response{
				Message: "Created role administrator",
				Status:  http.StatusCreated,
			},
		},
		{
			name:     "fail",
			expected: fmt.Errorf("error"),
			input:    "administrator",
			output: &types.Response{
				Message: "error",
				Status:  http.StatusInternalServerError,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			upr := new(Role)
			upr.Name = test.input
			payload, _ := json.Marshal(upr)

			req := httptest.NewRequest(http.MethodPost, "/api/v0/roles", bytes.NewReader(payload))

			mockService.EXPECT().CreateRole(gomock.Any(), "anonymous", test.input).Return(test.expected)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && len(rr.Data) != 0 {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestHandleCreateBadRoleFormat(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "no permissions",
			expected: nil,
			input:    "administrator",
			output: &types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			req := httptest.NewRequest(http.MethodPost, "/api/v0/roles", nil)

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

			if res.StatusCode != test.output.Status {
				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
			}

			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
			type Response struct {
				Data    []string          `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if rr.Message != test.output.Message {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
			}

			if rr.Status != test.output.Status {
				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
			}

		})
	}
}

func TestRegisterValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)
	mockValidationRegistry := NewMockValidationRegistryInterface(ctrl)

	apiKey := "roles"
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
