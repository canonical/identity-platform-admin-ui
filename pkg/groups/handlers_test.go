// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

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
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_validation.go -source=../../internal/validation/registry.go

// + http :8000/api/v0/groups X-Authorization:c2hpcHBlcml6ZXI=
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
//     "message": "List of groups",
//     "status": 200
// }

func TestHandleList(t *testing.T) {
	type expected struct {
		err    error
		groups []string
	}

	tests := []struct {
		name     string
		expected expected
		output   *types.Response
	}{
		{
			name: "empty result",
			expected: expected{
				groups: []string{},
				err:    nil,
			},
			output: &types.Response{
				Data:    []string{},
				Message: "List of groups",
				Status:  http.StatusOK,
			},
		},
		{
			name: "error",
			expected: expected{
				groups: []string{},
				err:    fmt.Errorf("error"),
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
				groups: []string{"global", "administrator", "viewer"},
				err:    nil,
			},

			output: &types.Response{
				Data:    []string{"global", "administrator", "viewer"},
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

			req := httptest.NewRequest(http.MethodGet, "/api/v0/groups", nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().ListGroups(gomock.Any(), gomock.Any()).Return(test.expected.groups, test.expected.err)

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

// + http :8000/api/v0/groups/administrator X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 77
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:32 GMT

// {
//     "_meta": null,
//     "data": [
//         "administrator"
//     ],
//     "message": "Group detail",
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
			name:     "unknown group",
			input:    "unknown",
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "group does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Data: []Group{{
					ID:   "administrator",
					Name: "administrator",
				}},
				Message: "Group detail",
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

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/groups/%s", test.input), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().GetGroup(gomock.Any(), gomock.Any(), test.input).DoAndReturn(
				func(context.Context, string, string) (*Group, error) {
					if test.expected != nil {
						return nil, test.expected
					}

					return &Group{
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
				Data    []Group           `json:"data"`
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
			name:     "unknown group",
			input:    "unknown",
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "use POST /api/v0/groups/unknown/entitlements to assign permissions",
				Status:  http.StatusNotImplemented,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Message: "use POST /api/v0/groups/administrator/entitlements to assign permissions",
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

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/groups/%s", test.input), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

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

// + http :8000/api/v0/groups/administrator/entitlements X-Authorization:c2hpcHBlcml6ZXI=
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

			groupID := "administrator"
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/groups/%s/entitlements", groupID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockService.EXPECT().ListPermissions(gomock.Any(), groupID, map[string]string{}).Return(test.expected.permissions, test.expected.cTokens, nil)

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

// + http :8000/api/v0/groups/administrator/roles X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 87
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:35 GMT

// {
//     "_meta": null,
//     "data": [
//         "viewer",
//         "devops",
//     ],
//     "message": "List of roles",
//     "status": 200
// }

func TestHandleListRolesSuccess(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		output   *types.Response
	}{
		{
			name:     "no roles",
			expected: []string{},
			output: &types.Response{
				Data:    []string{},
				Message: "List of roles",
				Status:  http.StatusOK,
			},
		},
		{
			name: "full roles",
			expected: []string{
				"viewer",
				"devops",
			},
			output: &types.Response{
				Data: []string{
					"viewer",
					"devops",
				},
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

			groupID := "administrator"
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/groups/%s/roles", groupID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().ListRoles(gomock.Any(), groupID).Return(test.expected, nil)

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

// + http DELETE :8000/api/v0/groups/administrator/entitlements/can_edit::client:okta X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 116
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Removed permission can_edit::client:okta for group administrator",
//	    "status": 200
//	}

func TestHandleRemovePermissionBadPermissionFormat(t *testing.T) {
	type input struct {
		groupID      string
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
				groupID:      "administrator",
				permissionID: "can_edit-something-wrong:okta",
			},
			expected: fmt.Errorf("group does not exist"),
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

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/groups/%s/entitlements/%s", test.input.groupID, test.input.permissionID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

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
		groupID      string
		permissionID string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name: "unknown group",
			input: input{
				groupID:      "unknown",
				permissionID: "can_edit::client::okta",
			},
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "group does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name: "found",
			input: input{
				groupID:      "administrator",
				permissionID: "can_edit::client:okta",
			},
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Removed permission can_edit::client:okta for group administrator",
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

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/groups/%s/entitlements/%s", test.input.groupID, test.input.permissionID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().RemovePermissions(
				gomock.Any(),
				test.input.groupID,
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

// + http PATCH :8000/api/v0/groups/administrator/entitlements 'permissions:=[{"relation":"can_delete","object":"scheme:superman"},{"relation":"can_view","object":"client:aws"}]' X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 201 Created
// Content-Length: 95
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:34 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Updated permissions for group administrator",
//	    "status": 201
//	}
func TestHandleAssignPermissions(t *testing.T) {
	type input struct {
		permissions []Permission
		groupID     string
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
				groupID: "administrator",
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
				Message: "Updated permissions for group administrator",
				Status:  http.StatusCreated,
			},
		},
		{
			name:     "multiple permissions with error",
			expected: fmt.Errorf("error"),
			input: input{
				groupID: "administrator",
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

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/groups/%s/entitlements", test.input.groupID), bytes.NewReader(payload))
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().AssignPermissions(gomock.Any(), test.input.groupID, test.input.permissions).Return(test.expected)

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

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/groups/%s/entitlements", test.input), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

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

// + http DELETE :8000/api/v0/groups/viewer X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 72
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:36 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Deleted group viewer",
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
			name:     "unknown group",
			input:    "unknown",
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "group does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Deleted group administrator",
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

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/groups/%s", test.input), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().DeleteGroup(
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
				Message: "Created group administrator",
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

			upr := new(Group)
			upr.Name = test.input
			payload, _ := json.Marshal(upr)

			req := httptest.NewRequest(http.MethodPost, "/api/v0/groups", bytes.NewReader(payload))
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			var group *Group = nil
			if test.expected == nil {
				group = &Group{ID: test.input, Name: test.input}
			}
			mockService.EXPECT().CreateGroup(gomock.Any(), gomock.Any(), test.input).Return(group, test.expected)

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
				Data    []Group           `json:"data"`
				Message string            `json:"message"`
				Status  int               `json:"status"`
				Meta    *types.Pagination `json:"_meta"`
			}

			rr := new(Response)

			if err := json.Unmarshal(data, rr); err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if test.expected == nil && (len(rr.Data) != 1 || rr.Data[0].Name != test.input || rr.Data[0].ID != test.input) {
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

			req := httptest.NewRequest(http.MethodPost, "/api/v0/groups", nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

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

// + http DELETE :8000/api/v0/groups/administrator/identities/joe X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 116
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Removed identity joe for group administrator",
//	    "status": 200
//	}
func TestHandleRemoveIdentities(t *testing.T) {
	type input struct {
		groupID    string
		identityID string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name: "unknown group",
			input: input{
				groupID:    "unknown",
				identityID: "joe",
			},
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "group does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name: "found",
			input: input{
				groupID:    "administrator",
				identityID: "joe",
			},
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Removed identity joe for group administrator",
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

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/groups/%s/identities/%s", test.input.groupID, test.input.identityID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().RemoveIdentities(
				gomock.Any(),
				test.input.groupID,
				test.input.identityID,
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

// + http PATCH :8000/api/v0/groups/administrator/identities 'identities:=["joe","susan"]' X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 201 Created
// Content-Length: 95
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:34 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Updated identities for group administrator",
//	    "status": 201
//	}
func TestHandleAssignIdentities(t *testing.T) {
	type input struct {
		identities []string
		groupID    string
	}

	tests := []struct {
		name             string
		input            input
		expectedCheck    bool
		expectedCheckErr error
		expected         error
		output           *types.Response
	}{
		{
			name:             "multiple identities",
			expectedCheck:    true,
			expectedCheckErr: nil,
			expected:         nil,
			input: input{
				groupID: "administrator",
				identities: []string{
					"joe", "susan", "dummy",
				},
			},
			output: &types.Response{
				Message: "Updated identities for group administrator",
				Status:  http.StatusCreated,
			},
		},
		{
			name:             "multiple identities cannot be assigned error",
			expectedCheck:    false,
			expectedCheckErr: nil,
			expected:         nil,
			input: input{
				groupID: "administrator",
				identities: []string{
					"joe", "susan", "dummy",
				},
			},
			output: &types.Response{
				Message: "user test-user is not allowed to assign specified identities",
				Status:  http.StatusForbidden,
			},
		},
		{
			name:             "multiple identities can be assigned then error",
			expectedCheck:    true,
			expectedCheckErr: nil,
			expected:         fmt.Errorf("error"),
			input: input{
				groupID: "administrator",
				identities: []string{
					"joe", "susan", "dummy",
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

			upr := new(UpdateIdentitiesRequest)
			upr.Identities = test.input.identities
			payload, _ := json.Marshal(upr)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/groups/%s/identities", test.input.groupID), bytes.NewReader(payload))
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().CanAssignIdentities(gomock.Any(), "test-user", test.input.identities).Return(test.expectedCheck, test.expectedCheckErr)
			if test.expectedCheck {
				mockService.EXPECT().AssignIdentities(gomock.Any(), test.input.groupID, test.input.identities).Return(test.expected)
			}

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

func TestHandleAssignIdentitiesBadPermissionFormat(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected error
		output   *types.Response
	}{
		{
			name:     "no identities",
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

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/groups/%s/identities", test.input), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

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

// + http PATCH :8000/api/v0/groups/administrator/roles 'roles:=["admin","viewer"]' X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 201 Created
// Content-Length: 95
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:34 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Updated roles for group administrator",
//	    "status": 201
//	}
func TestHandleAssignRoles(t *testing.T) {
	type input struct {
		roles   []string
		groupID string
	}

	tests := []struct {
		name             string
		input            input
		expectedCheck    bool
		expectedCheckErr error
		expected         error
		output           *types.Response
	}{
		{
			name:             "multiple roles",
			expectedCheck:    true,
			expectedCheckErr: nil,
			expected:         nil,
			input: input{
				groupID: "administrator",
				roles: []string{
					"viewer", "writer",
				},
			},
			output: &types.Response{
				Message: "Updated roles for group administrator",
				Status:  http.StatusCreated,
			},
		},
		{
			name:             "multiple roles cannot be assigned error",
			expectedCheck:    false,
			expectedCheckErr: nil,
			expected:         nil,
			input: input{
				groupID: "administrator",
				roles: []string{
					"viewer", "writer",
				},
			},
			output: &types.Response{
				Message: "user test-user is not allowed to assign specified roles",
				Status:  http.StatusForbidden,
			},
		},
		{
			name:             "multiple roles can be assigned then error",
			expectedCheck:    true,
			expectedCheckErr: nil,
			expected:         fmt.Errorf("error"),
			input: input{
				groupID: "administrator",
				roles: []string{
					"viewer", "writer",
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

			upr := new(UpdateRolesRequest)
			upr.Roles = test.input.roles
			payload, _ := json.Marshal(upr)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v0/groups/%s/roles", test.input.groupID), bytes.NewReader(payload))
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().CanAssignRoles(gomock.Any(), "test-user", test.input.roles).Return(test.expectedCheck, test.expectedCheckErr)
			if test.expectedCheck {
				mockService.EXPECT().AssignRoles(gomock.Any(), test.input.groupID, test.input.roles).Return(test.expected)
			}

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

// + http DELETE :8000/api/v0/groups/administrator/roles/viewer X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 116
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

//	{
//	    "_meta": null,
//	    "data": null,
//	    "message": "Removed role viewer for group administrator",
//	    "status": 200
//	}
func TestHandleRemoveRole(t *testing.T) {
	type input struct {
		groupID string
		roleID  string
	}

	tests := []struct {
		name     string
		input    input
		expected error
		output   *types.Response
	}{
		{
			name: "unknown group",
			input: input{
				groupID: "unknown",
				roleID:  "viewer",
			},
			expected: fmt.Errorf("group does not exist"),
			output: &types.Response{
				Message: "group does not exist",
				Status:  http.StatusInternalServerError,
			},
		},
		{
			name: "found",
			input: input{
				groupID: "administrator",
				roleID:  "viewer",
			},
			expected: nil,
			output: &types.Response{
				Status:  http.StatusOK,
				Message: "Removed role viewer from group administrator",
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

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/groups/%s/roles/%s", test.input.groupID, test.input.roleID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockService.EXPECT().RemoveRoles(
				gomock.Any(),
				test.input.groupID,
				test.input.roleID,
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

// + http :8000/api/v0/groups/administrator/identities X-Authorization:c2hpcHBlcml6ZXI=
// HTTP/1.1 200 OK
// Content-Length: 156
// Content-Type: application/json
// Date: Tue, 20 Feb 2024 22:10:33 GMT

// {
//     "_meta": null,
//     "data": [
//        "user:joe",
//        "user:susan",
//        "user:dummy",
//     ],
//     "message": "List of entitlements",
//     "status": 200
// }

func TestHandleListIdentitiesSuccess(t *testing.T) {
	type expected struct {
		identities []string
		cToken     string
	}

	tests := []struct {
		name     string
		expected expected
		output   *types.Response
	}{
		{
			name:     "no identities",
			expected: expected{identities: []string{}},
			output: &types.Response{
				Data:    []string{},
				Message: "List of identities",
				Status:  http.StatusOK,
			},
		},
		{
			name: "full identities",
			expected: expected{
				identities: []string{
					"user:joe", "user:susan", "user:dummy",
				},
				cToken: "test",
			},
			output: &types.Response{
				Data: []string{
					"user:joe", "user:susan", "user:dummy",
				},
				Message: "List of identities",
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

			groupID := "administrator"
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/groups/%s/identities", groupID), nil)
			req = req.WithContext(authentication.PrincipalContext(req.Context(), &authentication.UserPrincipal{Email: "test-user"}))

			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockService.EXPECT().ListIdentities(gomock.Any(), groupID, "").Return(test.expected.identities, test.expected.cToken, nil)

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

			if test.expected.cToken != "" {
				if err != nil {
					t.Errorf("expected continuation token in headers")
				}

				tokens := map[string]string{}

				_ = json.Unmarshal(tokenMap, &tokens)

				if !reflect.DeepEqual(tokens[GROUP_TOKEN_KEY], test.expected.cToken) {
					t.Errorf("expected continuation token to match: %v - %v", tokens[GROUP_TOKEN_KEY], test.expected.cToken)
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

func TestRegisterValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockService := NewMockServiceInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockValidationRegistry := NewMockValidationRegistryInterface(ctrl)

	apiKey := "groups"
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
