// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
)

//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package pool -destination ../../internal/pool/mock_pool.go -source=../../internal/pool/interfaces.go

func setupMockSubmit(wp *pool.MockWorkerPoolInterface, resultsChan chan *pool.Result[any]) (*gomock.Call, chan *pool.Result[any]) {
	key := uuid.New()
	var internalResultsChannel chan *pool.Result[any]

	call := wp.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Do(
		func(command any, results chan *pool.Result[any], wg *sync.WaitGroup) {
			var value any = true

			switch commandFunc := command.(type) {
			case func():
				commandFunc()
			case func() any:
				value = commandFunc()
			}

			result := pool.NewResult[any](key, value)
			results <- result
			if resultsChan != nil {
				resultsChan <- result
			}

			wg.Done()

			internalResultsChannel = results
		},
	).Return(key.String(), nil)

	return call, internalResultsChannel
}

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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

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

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.GetRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().Check(gomock.Any(), fmt.Sprintf("user:%s", test.input.user), "can_view", fmt.Sprintf("role:%s", test.input.role)).Return(test.expected.check, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			role, err := svc.GetRole(context.Background(), test.input.user, test.input.role)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && test.expected.check && role.ID != test.input.role {
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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.CreateRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					ps = append(
						ps,
						*ofga.NewTuple(fmt.Sprintf("user:%s", test.input.user), ASSIGNEE_RELATION, fmt.Sprintf("role:%s", test.input.role)),
						*ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("role:%s", test.input.role)),
						*ofga.NewTuple(fmt.Sprintf("user:%s", test.input.user), CAN_VIEW_RELATION, fmt.Sprintf("role:%s", test.input.role)),
					)

					if !reflect.DeepEqual(ps, tuples) {
						t.Errorf("expected tuples to be %v got %v", ps, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.CreateRole(context.Background(), test.input.user, test.input.role)

			if test.expected != nil && err != test.expected {
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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)
			for i := 0; i < 7; i++ {
				setupMockSubmit(workerPool, nil)
			}

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.DeleteRole").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.removePermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.removeDirectAssociations").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			pTypes := []string{"role", "group", "identity", "scheme", "provider", "client"}
			directRelations := []string{"privileged", "assignee", "can_create", "can_delete", "can_edit", "can_view"}

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

			for _, relation := range directRelations {
				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), "", relation, fmt.Sprintf("role:%s", test.input), "").Times(1).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										"user:test", ASSIGNEE_RELATION, object,
									),
									time.Now(),
								),
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										"group:test#member", ASSIGNEE_RELATION, object,
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
				).Times(12).DoAndReturn(
					func(ctx context.Context, tuples ...ofga.Tuple) error {

						switch len(tuples) {
						case 1:
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
						case 2:

							for _, tuple := range tuples {
								if tuple.User != "user:test" && tuple.User != "group:test#member" {
									t.Errorf("expected user to be one of %v got %v", []string{"user:test", "group:test#member"}, tuple.User)
								}

								if tuple.Relation != ASSIGNEE_RELATION {
									t.Errorf("expected relation to be of %v got %v", ASSIGNEE_RELATION, tuple.Relation)
								}

								if tuple.Object != fmt.Sprintf("role:%s", test.input) {
									t.Errorf("expected object to be one of %v got %v", fmt.Sprintf("role:%s", test.input), tuple.Object)
								}
							}

						default:
							t.Errorf("too many tuples")
						}

						return nil
					},
				)
			} else {
				// TODO @shipperizer fix this so that we can pin it down to the error case only
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			}

			_ = svc.DeleteRole(context.Background(), test.input)

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

			mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
			workerPool := pool.NewMockWorkerPoolInterface(ctrl)
			for i := 0; i < 6; i++ {
				setupMockSubmit(workerPool, nil)
			}

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.ListPermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.listPermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			pTypes := []string{"role", "group", "identity", "scheme", "provider", "client"}
			expCTokens := map[string]string{
				"role":     "",
				"group":    "",
				"identity": "",
				"scheme":   "",
				"provider": "",
				"client":   "",
			}

			expPermissions := []string{
				"can_edit::role:test",
				"can_edit::group:test",
				"can_edit::identity:test",
				"can_edit::scheme:test",
				"can_edit::provider:test",
				"can_edit::client:test",
			}

			calls := []*gomock.Call{}

			for _, _ = range pTypes {
				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							if user != fmt.Sprintf("role:%s#%s", test.input.role, ASSIGNEE_RELATION) {
								t.Errorf("wrong user parameter expected %s got %s", fmt.Sprintf("role:%s#%s", test.input.role, ASSIGNEE_RELATION), user)
							}

							if object == "role:" && continuationToken != "test" {
								t.Errorf("missing continuation token %s", test.input.cTokens["roles"])
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "can_edit", fmt.Sprintf("%stest", object),
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

			if test.expected != nil {
				// TODO @shipperizer fix this so that we can pin it down to the error case only
				mockLogger.EXPECT().Error(gomock.Any()).MinTimes(0).MaxTimes(12)
				mockLogger.EXPECT().Errorf(gomock.Any()).MaxTimes(12)
			}

			gomock.InAnyOrder(calls)
			permissions, cTokens, err := svc.ListPermissions(context.Background(), test.input.role, test.input.cTokens)

			if err != nil && test.expected == nil {
				t.Fatalf("expected error to be silenced and return nil got %v instead", err)
			}

			sort.Strings(permissions)
			sort.Strings(expPermissions)

			if err == nil && test.expected == nil && !reflect.DeepEqual(permissions, expPermissions) {
				t.Fatalf("expected permissions to be %v got %v", expPermissions, permissions)
			}

			if err == nil && test.expected == nil && !reflect.DeepEqual(cTokens, expCTokens) {
				t.Fatalf("expected continuation tokens to be %v got %v", expCTokens, cTokens)
			}
		})
	}
}

func TestServiceAssignPermissions(t *testing.T) {
	type input struct {
		role        string
		permissions []Permission
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
				permissions: []Permission{
					{Relation: "can_delete", Object: "role:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions",
			input: input{
				role: "administrator",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.AssignPermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *ofga.NewTuple(fmt.Sprintf("role:%s#%s", test.input.role, ASSIGNEE_RELATION), p.Relation, p.Object))
					}

					if !reflect.DeepEqual(ps, tuples) {
						t.Errorf("expected tuples to be %v got %v", ps, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.AssignPermissions(context.Background(), test.input.role, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceRemovePermissions(t *testing.T) {
	type input struct {
		role        string
		permissions []Permission
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
				permissions: []Permission{
					{Relation: "can_delete", Object: "role:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions",
			input: input{
				role: "administrator",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
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

			workerPool := pool.NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "roles.Service.RemovePermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *ofga.NewTuple(fmt.Sprintf("role:%s#%s", test.input.role, ASSIGNEE_RELATION), p.Relation, p.Object))
					}

					if !reflect.DeepEqual(ps, tuples) {
						t.Errorf("expected tuples to be %v got %v", ps, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.RemovePermissions(context.Background(), test.input.role, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}
