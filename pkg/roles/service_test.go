// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package roles

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestServiceListRoles(t *testing.T) {
	type expected struct {
		err   error
		roles []string
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "empty result",
			input: "administrator",
			expected: expected{
				roles: []string{},
				err:   nil,
			},
		},
		{
			name:  "error",
			input: "administrator",
			expected: expected{
				roles: []string{},
				err:   fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: "administrator",
			expected: expected{
				roles: []string{"global", "administrator", "viewer"},
				err:   nil,
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
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.ListRoles").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ListObjects(gomock.Any(), fmt.Sprintf("user:%s", test.input), "can_view", "role").Return(test.expected.roles, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			roles, err := svc.ListRoles(context.Background(), test.input)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(roles, test.expected.roles) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.roles, roles)
			}
		})
	}
}

func TestServiceListRoleGroups(t *testing.T) {
	type expected struct {
		err    error
		tuples []string
		token  string
	}

	type input struct {
		role  string
		token string
	}

	tests := []struct {
		name     string
		input    input
		expected expected
		output   []string
	}{
		{
			name: "empty result",
			input: input{
				role: "administrator",
			},
			expected: expected{
				tuples: []string{},
				token:  "",
				err:    nil,
			},
			output: []string{},
		},
		{
			name: "error",
			input: input{
				role: "administrator",
			},
			expected: expected{
				tuples: []string{},
				token:  "",
				err:    fmt.Errorf("error"),
			},
		},
		{
			name: "full result without token",
			input: input{
				role: "administrator",
			},
			expected: expected{
				tuples: []string{
					"group:c-level#member",
					"group:it-admin#member",
					"user:joe",
					"user:test",
				},
				token: "test",
				err:   nil,
			},
			output: []string{
				"group:c-level#member",
				"group:it-admin#member",
			},
		},
		{
			name: "full result with token",
			input: input{
				role:  "administrator",
				token: "test",
			},
			expected: expected{
				tuples: []string{
					"group:c-level#member",
					"group:it-admin#member",
				},
				token: "",
				err:   nil,
			},
			output: []string{
				"group:c-level#member",
				"group:it-admin#member",
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
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			r := new(client.ClientReadResponse)

			tuples := []openfga.Tuple{}
			for _, t := range test.expected.tuples {
				tuples = append(
					tuples,
					*openfga.NewTuple(
						*openfga.NewTupleKey(
							t, ASSIGNEE_RELATION, fmt.Sprintf("role:%s", test.input.role),
						),
						time.Now(),
					),
				)
			}

			r.SetContinuationToken(test.expected.token)
			r.SetTuples(tuples)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.ListRoleGroups").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), "", ASSIGNEE_RELATION, fmt.Sprintf("role:%s", test.input.role), test.input.token).Return(r, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			groups, token, err := svc.ListRoleGroups(context.Background(), test.input.role, test.input.token)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && token != test.expected.token {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.token, token)
			}

			if test.expected.err == nil && !reflect.DeepEqual(groups, test.output) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output, groups)
			}
		})
	}
}

func TestServiceGetRole(t *testing.T) {
	type expected struct {
		err   error
		check bool
	}

	type input struct {
		role string
		user string
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "not found",
			input: input{
				role: "administrator",
				user: "admin",
			},
			expected: expected{
				check: false,
				err:   nil,
			},
		},
		{
			name: "error",
			input: input{
				role: "administrator",
				user: "admin",
			},
			expected: expected{
				check: false,
				err:   fmt.Errorf("error"),
			},
		},
		{
			name: "found",
			input: input{
				role: "administrator",
				user: "admin",
			},
			expected: expected{
				check: true,
				err:   nil,
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
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.GetRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().Check(gomock.Any(), fmt.Sprintf("user:%s", test.input.user), "can_view", fmt.Sprintf("role:%s", test.input.role)).Return(test.expected.check, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			role, err := svc.GetRole(context.Background(), test.input.user, test.input.role)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && test.expected.check && role != test.input.role {
				t.Errorf("invalid result, expected: %v, got: %v", test.input.role, role)
			}
		})
	}
}

func TestServiceCreateRole(t *testing.T) {
	type input struct {
		role string
		user string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				role: "administrator",
				user: "admin",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "found",
			input: input{
				role: "administrator",
				user: "admin",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.CreateRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuple(gomock.Any(), fmt.Sprintf("user:%s", test.input.user), ASSIGNEE_RELATION, fmt.Sprintf("role:%s", test.input.role)).Return(test.expected)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			} else {
				mockOpenFGA.EXPECT().WriteTuple(gomock.Any(), authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("role:%s", test.input.role)).Return(test.expected)
			}

			err := svc.CreateRole(context.Background(), test.input.user, test.input.role)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

// TODO @shipperizer split this test in 2, test only specific ofga client calls in each
func TestServiceDeleteRole(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "error",
			input:    "administrator",
			expected: fmt.Errorf("error"),
		},
		{
			name:     "found",
			input:    "administrator",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.DeleteRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.removePermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			pTypes := []string{"role", "group", "identity", "scheme", "provider", "client"}

			calls := []*gomock.Call{}

			for _, pType := range pTypes {

				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), fmt.Sprintf("role:%s#%s", test.input, ASSIGNEE_RELATION), "", fmt.Sprintf("%s:", pType), "").Times(1).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "can_edit", fmt.Sprintf("%s:test", pType),
									),
									time.Now(),
								),
							}

							r := new(client.ClientReadResponse)
							r.SetContinuationToken("")
							r.SetTuples(tuples)

							return r, nil
						},
					),
				)

			}

			if test.expected == nil {
				mockOpenFGA.EXPECT().DeleteTuples(
					gomock.Any(),
					gomock.Any(),
				).Times(7).DoAndReturn(
					func(ctx context.Context, tuples ...ofga.Tuple) error {
						if len(tuples) != 1 {
							t.Errorf("too many tuples")
						}

						tuple := tuples[0]

						if tuple.User != fmt.Sprintf("role:%s#%s", test.input, ASSIGNEE_RELATION) && tuple.User != authorization.ADMIN_PRIVILEGE {
							t.Errorf("expected user to be one of %v got %v", []string{fmt.Sprintf("role:%s#%s", test.input, ASSIGNEE_RELATION), authorization.ADMIN_PRIVILEGE}, tuple.User)
						}

						if tuple.Relation != "privileged" && tuple.Relation != "can_edit" {
							t.Errorf("expected relation to be one of %v got %v", []string{"privileged", "can_edit"}, tuple.Relation)
						}

						if tuple.Object != fmt.Sprintf("role:%s", test.input) && !strings.HasSuffix(tuple.Object, ":test") {
							t.Errorf("expected object to be one of %v got %v", []string{fmt.Sprintf("role:%s", test.input), "<*>:test"}, tuple.Object)
						}

						return nil
					},
				)
			} else {
				// TODO @shipperizer fix this so that we can pin it down to the error case only
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockOpenFGA.EXPECT().DeleteTuples(
					gomock.Any(),
					*ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("role:%s", test.input)),
				).Return(test.expected)
			}

			gomock.InAnyOrder(calls)
			err := svc.DeleteRole(context.Background(), test.input)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceListPermissions(t *testing.T) {
	type input struct {
		role    string
		cTokens map[string]string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				role: "administrator",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "found",
			input: input{
				role: "administrator",
				cTokens: map[string]string{
					"role": "test",
				},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)

			svc := NewService(mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.ListPermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.listPermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			pTypes := []string{"role", "group", "identity", "scheme", "provider", "client"}
			expPermissions := make([]string, 0)
			expCTokens := map[string]string{
				"role":     "",
				"group":    "",
				"identity": "",
				"scheme":   "",
				"provider": "",
				"client":   "",
			}

			calls := []*gomock.Call{}

			for _, pType := range pTypes {
				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), fmt.Sprintf("role:%s#%s", test.input.role, ASSIGNEE_RELATION), "", fmt.Sprintf("%s:", pType), test.input.cTokens[pType]).Times(1).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {

							if test.expected != nil {
								return nil, test.expected
							}

							if object == "role:" && continuationToken != "test" {
								t.Errorf("missing continuation token %s", test.input.cTokens["roles"])
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "can_edit", fmt.Sprintf("%s:test", pType),
									),
									time.Now(),
								),
							}

							r := new(client.ClientReadResponse)
							r.SetContinuationToken("")
							r.SetTuples(tuples)

							expPermissions = append(
								expPermissions,
								authorization.NewUrn(tuples[0].Key.Relation, tuples[0].Key.Object).ID(),
							)

							return r, nil
						},
					),
				)
			}

			if test.expected != nil {
				// TODO @shipperizer fix this so that we can pin it down to the error case only
				mockLogger.EXPECT().Error(gomock.Any()).Times(12)
			}

			gomock.InAnyOrder(calls)
			permissions, cTokens, err := svc.ListPermissions(context.Background(), test.input.role, test.input.cTokens)

			if err != nil && test.expected != nil {
				t.Errorf("expected error to be silenced and return nil got %v instead", err)
			}

			if err == nil && !reflect.DeepEqual(permissions, expPermissions) {
				t.Errorf("expected permissions to be %v got %v", expPermissions, permissions)
			}

			if err == nil && !reflect.DeepEqual(cTokens, expCTokens) {
				t.Errorf("expected continuation tokens to be %v got %v", expCTokens, cTokens)
			}
		})
	}
}

// // + http :8000/api/v0/roles/administrator X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 200 OK
// // Content-Length: 77
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:32 GMT

// // {
// //     "_meta": null,
// //     "data": [
// //         "administrator"
// //     ],
// //     "message": "Rule detail",
// //     "status": 200
// // }

// func TestHandleDetail(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		input    string
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "unknown role",
// 			input:    "unknown",
// 			expected: fmt.Errorf("role does not exist"),
// 			output: &types.Response{
// 				Message: "role does not exist",
// 				Status:  http.StatusInternalServerError,
// 			},
// 		},
// 		{
// 			name:     "found",
// 			input:    "administrator",
// 			expected: nil,
// 			output: &types.Response{
// 				Data:    []string{"administrator"},
// 				Message: "Rule detail",
// 				Status:  http.StatusOK,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s", test.input), nil)

// 			mockService.EXPECT().GetRole(gomock.Any(), "anonymous", test.input).DoAndReturn(
// 				func(context.Context, string, string) (string, error) {
// 					if test.expected != nil {
// 						return "", test.expected
// 					}

// 					return test.input, nil

// 				},
// 			)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && !reflect.DeepEqual(rr.Data, test.output.Data) {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// // + http :8000/api/v0/roles/administrator/entitlements X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 200 OK
// // Content-Length: 156
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:33 GMT

// // {
// //     "_meta": null,
// //     "data": [
// //        "can_view::client:github-canonical",
// //        "can_delete::client:okta",
// //        "can_edit::client:okta"
// //     ],
// //     "message": "List of entitlements",
// //     "status": 200
// // }

// func TestHandleListPermissionsSuccess(t *testing.T) {
// 	type expected struct {
// 		permissions []string
// 		cTokens     map[string]string
// 	}

// 	tests := []struct {
// 		name     string
// 		expected expected
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "no permissions",
// 			expected: expected{permissions: []string{}},
// 			output: &types.Response{
// 				Data:    []string{},
// 				Message: "List of entitlements",
// 				Status:  http.StatusOK,
// 			},
// 		},
// 		{
// 			name: "full permissions",
// 			expected: expected{
// 				permissions: []string{
// 					"can_view::client:github-canonical",
// 					"can_delete::client:okta",
// 					"can_edit::client:okta",
// 				},
// 				cTokens: map[string]string{"client": "test"},
// 			},
// 			output: &types.Response{
// 				Data: []string{
// 					"can_view::client:github-canonical",
// 					"can_delete::client:okta",
// 					"can_edit::client:okta",
// 				},
// 				Message: "List of entitlements",
// 				Status:  http.StatusOK,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			roleID := "administrator"
// 			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s/entitlements", roleID), nil)

// 			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
// 			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

// 			mockService.EXPECT().ListPermissions(gomock.Any(), roleID, map[string]string{}).Return(test.expected.permissions, test.expected.cTokens, nil)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != http.StatusOK {
// 				t.Errorf("expected HTTP status code 200 got %v", res.StatusCode)
// 			}

// 			tokenMap, err := base64.StdEncoding.DecodeString(res.Header.Get(types.PAGINATION_HEADER))

// 			if test.expected.cTokens != nil {
// 				if err != nil {
// 					t.Errorf("expected continuation token in headers")
// 				}

// 				tokens := map[string]string{}

// 				_ = json.Unmarshal(tokenMap, &tokens)

// 				if !reflect.DeepEqual(tokens, test.expected.cTokens) {
// 					t.Errorf("expected continuation tokens to match: %v - %v", tokens, test.expected.cTokens)
// 				}
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if !reflect.DeepEqual(rr.Data, test.output.Data) {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// // + http :8000/api/v0/roles/administrator/groups X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 200 OK
// // Content-Length: 87
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:35 GMT

// // {
// //     "_meta": null,
// //     "data": [
// //         "group:c-level#member"
// //     ],
// //     "message": "List of groups",
// //     "status": 200
// // }

// func TestHandleListRoleGroupsSuccess(t *testing.T) {
// 	type expected struct {
// 		groups  []string
// 		cTokens map[string]string
// 	}

// 	tests := []struct {
// 		name     string
// 		expected expected
// 		output   *types.Response
// 	}{
// 		{
// 	name:     "no groups",
// 	expected: expected{groups: []string{}},
// 	output: &types.Response{
// 		Data:    []string{},
// 		Message: "List of groups",
// 		Status:  http.StatusOK,
// 	},
// },
// {
// 	name: "full groups",
// 	expected: expected{
// 		groups: []string{
// 			"group:c-level#member",
// 			"group:it-admin#member",
// 		},
// 		cTokens: map[string]string{"roles": "test"},
// 	},
// 	output: &types.Response{
// 		Data: []string{
// 			"group:c-level#member",
// 			"group:it-admin#member",
// 		},
// 		Message: "List of groups",
// 		Status:  http.StatusOK,
// 	},
// },
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			roleID := "administrator"
// 			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v0/roles/%s/groups", roleID), nil)

// 			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.LoadFromRequest").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
// 			mockTracer.EXPECT().Start(gomock.Any(), "types.TokenPaginator.PaginationHeader").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

// 			mockService.EXPECT().ListRoleGroups(gomock.Any(), roleID, "").Return(test.expected.groups, test.expected.cTokens["roles"], nil)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != http.StatusOK {
// 				t.Errorf("expected HTTP status code 200 got %v", res.StatusCode)
// 			}

// 			tokenMap, err := base64.StdEncoding.DecodeString(res.Header.Get(types.PAGINATION_HEADER))

// 			if test.expected.cTokens != nil {
// 				if err != nil {
// 					t.Errorf("expected continuation token in headers")
// 				}

// 				tokens := map[string]string{}

// 				_ = json.Unmarshal(tokenMap, &tokens)

// 				if !reflect.DeepEqual(tokens, test.expected.cTokens) {
// 					t.Errorf("expected continuation tokens to match: %v - %v", tokens, test.expected.cTokens)
// 				}
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if !reflect.DeepEqual(rr.Data, test.output.Data) {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// // + http DELETE :8000/api/v0/roles/administrator/entitlements/can_edit::client:okta X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 200 OK
// // Content-Length: 116
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:33 GMT

// //	{
// //	    "_meta": null,
// //	    "data": null,
// //	    "message": "Removed permission can_edit::client:okta for role administrator",
// //	    "status": 200
// //	}

// func TestHandleRemovePermissionBadPermissionFormat(t *testing.T) {
// 	type input struct {
// 		roleID       string
// 		permissionID string
// 	}

// 	tests := []struct {
// 		name     string
// 		input    input
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name: "wrong permission format",
// 			input: input{
// 				roleID:       "administrator",
// 				permissionID: "can_edit-something-wrong:okta",
// 			},
// 			expected: fmt.Errorf("role does not exist"),
// 			output: &types.Response{
// 				Message: "Error parsing entitlement ID",
// 				Status:  http.StatusBadRequest,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s/entitlements/%s", test.input.roleID, test.input.permissionID), nil)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && !reflect.DeepEqual(rr.Data, test.output.Data) {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// func TestHandleRemovePermission(t *testing.T) {
// 	type input struct {
// 		roleID       string
// 		permissionID string
// 	}

// 	tests := []struct {
// 		name     string
// 		input    input
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name: "unknown role",
// 			input: input{
// 				roleID:       "unknown",
// 				permissionID: "can_edit::client::okta",
// 			},
// 			expected: fmt.Errorf("role does not exist"),
// 			output: &types.Response{
// 				Message: "role does not exist",
// 				Status:  http.StatusInternalServerError,
// 			},
// 		},
// 		{
// 			name: "found",
// 			input: input{
// 				roleID:       "administrator",
// 				permissionID: "can_edit::client:okta",
// 			},
// 			expected: nil,
// 			output: &types.Response{
// 				Status:  http.StatusOK,
// 				Message: "Removed permission can_edit::client:okta for role administrator",
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s/entitlements/%s", test.input.roleID, test.input.permissionID), nil)

// 			mockService.EXPECT().RemovePermissions(
// 				gomock.Any(),
// 				test.input.roleID,
// 				Permission{
// 					Relation: strings.Split(test.input.permissionID, authorization.PERMISSION_SEPARATOR)[0],
// 					Object:   strings.Split(test.input.permissionID, authorization.PERMISSION_SEPARATOR)[1],
// 				},
// 			).Return(test.expected)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && len(rr.Data) != 0 {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// // + http PATCH :8000/api/v0/roles/administrator/entitlements 'permissions:=[{"relation":"can_delete","object":"scheme:superman"},{"relation":"can_view","object":"client:aws"}]' X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 201 Created
// // Content-Length: 95
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:34 GMT

// //	{
// //	    "_meta": null,
// //	    "data": null,
// //	    "message": "Updated permissions for role administrator",
// //	    "status": 201
// //	}
// func TestHandleAssignPermissions(t *testing.T) {
// 	type input struct {
// 		permissions []Permission
// 		roleID      string
// 	}

// 	tests := []struct {
// 		name     string
// 		input    input
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "multiple permissions",
// 			expected: nil,
// 			input: input{
// 				roleID: "administrator",
// 				permissions: []Permission{
// 					{
// 						Relation: "can_view",
// 						Object:   "client:github-canonical",
// 					},
// 					{
// 						Relation: "can_delete",
// 						Object:   "client:okta",
// 					},
// 					{
// 						Relation: "can_edit",
// 						Object:   "client:okta",
// 					},
// 				},
// 			},
// 			output: &types.Response{
// 				Message: "Updated permissions for role administrator",
// 				Status:  http.StatusCreated,
// 			},
// 		},
// 		{
// 			name:     "multiple permissions with error",
// 			expected: fmt.Errorf("error"),
// 			input: input{
// 				roleID: "administrator",
// 				permissions: []Permission{
// 					{
// 						Relation: "can_view",
// 						Object:   "client:github-canonical",
// 					},
// 					{
// 						Relation: "can_delete",
// 						Object:   "client:okta",
// 					},
// 					{
// 						Relation: "can_edit",
// 						Object:   "client:okta",
// 					},
// 				},
// 			},
// 			output: &types.Response{
// 				Message: "error",
// 				Status:  http.StatusInternalServerError,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			upr := new(UpdatePermissionsRequest)
// 			upr.Permissions = test.input.permissions
// 			payload, _ := json.Marshal(upr)

// 			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/roles/%s/entitlements", test.input.roleID), bytes.NewReader(payload))

// 			mockService.EXPECT().AssignPermissions(gomock.Any(), test.input.roleID, test.input.permissions).Return(test.expected)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && len(rr.Data) != 0 {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// func TestHandleAssignPermissionsBadPermissionFormat(t *testing.T) {

// 	tests := []struct {
// 		name     string
// 		input    string
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "no permissions",
// 			expected: nil,
// 			input:    "administrator",
// 			output: &types.Response{
// 				Message: "Error parsing JSON payload",
// 				Status:  http.StatusBadRequest,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v0/roles/%s/entitlements", test.input), nil)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// // + http DELETE :8000/api/v0/roles/viewer X-Authorization:c2hpcHBlcml6ZXI=
// // HTTP/1.1 200 OK
// // Content-Length: 72
// // Content-Type: application/json
// // Date: Tue, 20 Feb 2024 22:10:36 GMT

// //	{
// //	    "_meta": null,
// //	    "data": null,
// //	    "message": "Deleted role viewer",
// //	    "status": 200
// //	}
// func TestHandleRemove(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		input    string
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "unknown role",
// 			input:    "unknown",
// 			expected: fmt.Errorf("role does not exist"),
// 			output: &types.Response{
// 				Message: "role does not exist",
// 				Status:  http.StatusInternalServerError,
// 			},
// 		},
// 		{
// 			name:     "found",
// 			input:    "administrator",
// 			expected: nil,
// 			output: &types.Response{
// 				Status:  http.StatusOK,
// 				Message: "Deleted role administrator",
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v0/roles/%s", test.input), nil)

// 			mockService.EXPECT().DeleteRole(
// 				gomock.Any(),
// 				test.input,
// 			).Return(test.expected)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && len(rr.Data) != 0 {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// func TestHandleCreate(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		input    string
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "success",
// 			expected: nil,
// 			input:    "administrator",

// 			output: &types.Response{
// 				Message: "Created role administrator",
// 				Status:  http.StatusCreated,
// 			},
// 		},
// 		{
// 			name:     "fail",
// 			expected: fmt.Errorf("error"),
// 			input:    "administrator",
// 			output: &types.Response{
// 				Message: "error",
// 				Status:  http.StatusInternalServerError,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			upr := new(RoleRequest)
// 			upr.ID = test.input
// 			payload, _ := json.Marshal(upr)

// 			req := httptest.NewRequest(http.MethodPost, "/api/v0/roles", bytes.NewReader(payload))

// 			mockService.EXPECT().CreateRole(gomock.Any(), "anonymous", test.input).Return(test.expected)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if test.expected == nil && len(rr.Data) != 0 {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Data, rr.Data)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }

// func TestHandleCreateBadRoleFormat(t *testing.T) {

// 	tests := []struct {
// 		name     string
// 		input    string
// 		expected error
// 		output   *types.Response
// 	}{
// 		{
// 			name:     "no permissions",
// 			expected: nil,
// 			input:    "administrator",
// 			output: &types.Response{
// 				Message: "Error parsing JSON payload",
// 				Status:  http.StatusBadRequest,
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockLogger := NewMockLoggerInterface(ctrl)
// 			mockTracer := NewMockTracer(ctrl)
// 			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
// 			mockService := NewMockServiceInterface(ctrl)

// 			req := httptest.NewRequest(http.MethodPost, "/api/v0/roles", nil)

// 			w := httptest.NewRecorder()
// 			mux := chi.NewMux()
// 			NewAPI(mockService, mockTracer, mockMonitor, mockLogger).RegisterEndpoints(mux)

// 			mux.ServeHTTP(w, req)

// 			res := w.Result()
// 			defer res.Body.Close()
// 			data, err := io.ReadAll(res.Body)

// 			if err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if res.StatusCode != test.output.Status {
// 				t.Errorf("expected HTTP status code %v got %v", test.output.Status, res.StatusCode)
// 			}

// 			// duplicate types.Response attribute we care and assign the proper type instead of interface{}
// 			type Response struct {
// 				Data    []string          `json:"data"`
// 				Message string            `json:"message"`
// 				Status  int               `json:"status"`
// 				Meta    *types.Pagination `json:"_meta"`
// 			}

// 			rr := new(Response)

// 			if err := json.Unmarshal(data, rr); err != nil {
// 				t.Errorf("expected error to be nil got %v", err)
// 			}

// 			if rr.Message != test.output.Message {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Message, rr.Message)
// 			}

// 			if rr.Status != test.output.Status {
// 				t.Errorf("invalid result, expected: %v, got: %v", test.output.Status, rr.Status)
// 			}

// 		})
// 	}
// }
