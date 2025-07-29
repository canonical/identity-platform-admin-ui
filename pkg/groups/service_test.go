// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"

	"github.com/google/uuid"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	authz "github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
)

//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_openfga_interfaces.go -source=../../internal/openfga/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_pool.go -source=../../internal/pool/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_authentication.go -source=../authentication/interfaces.go

func setupMockSubmit(wp *MockWorkerPoolInterface, resultsChan chan *pool.Result[any]) (*gomock.Call, chan *pool.Result[any]) {
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

func TestServiceListGroups(t *testing.T) {
	type expected struct {
		err    error
		groups []string
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
				groups: []string{},
				err:    nil,
			},
		},
		{
			name:  "error",
			input: "administrator",
			expected: expected{
				groups: []string{},
				err:    fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: "administrator",
			expected: expected{
				groups: []string{"global", "administrator", "devops"},
				err:    nil,
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
			workerPool := NewMockWorkerPoolInterface(ctrl)
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.ListGroups").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockRepo.EXPECT().ListGroups(gomock.Any(), test.input, gomock.Any(), gomock.Any()).Times(1).Return(test.expected.groups, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			groups, err := svc.ListGroups(context.Background(), test.input)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(groups, test.expected.groups) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.groups, groups)
			}
		})
	}
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.ListRoles").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ListObjects(gomock.Any(), fmt.Sprintf("group:%s#%s", test.input, authz.MEMBER_RELATION), authz.ASSIGNEE_RELATION, "role").Return(test.expected.roles, test.expected.err)

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

func TestServiceListIdentities(t *testing.T) {
	type tuples struct {
		users  []string
		groups []string
	}

	type expected struct {
		errUsersListUsers  error
		errGroupsListUsers error
		tuples             tuples
	}

	tests := []struct {
		name     string
		group    string
		expected expected
		output   []string
	}{
		{
			name:  "empty result",
			group: "administrator",
			expected: expected{
				tuples: tuples{
					users:  []string{},
					groups: []string{},
				},
			},
			output: []string{},
		},
		{
			name:  "error-users-listusers",
			group: "administrator",
			expected: expected{
				tuples: tuples{
					users:  []string{},
					groups: []string{},
				},
				errUsersListUsers: fmt.Errorf("error"),
			},
		},
		{
			name:  "error-groups-listusers",
			group: "administrator",
			expected: expected{
				tuples: tuples{
					users:  []string{},
					groups: []string{},
				},
				errGroupsListUsers: fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			group: "administrator",
			expected: expected{
				tuples: tuples{
					users: []string{
						"user:joe",
						"user:test",
					},
					groups: []string{
						"group:group1",
					},
				},
			},
			output: []string{
				"user:joe",
				"user:test",
				"group:group1",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.ListIdentities").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockOpenFGA.EXPECT().ListUsers(gomock.Any(), "user", authz.MEMBER_RELATION, fmt.Sprintf("group:%s", test.group)).Return(test.expected.tuples.users, test.expected.errUsersListUsers)
			if test.expected.errUsersListUsers != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			} else {
				mockOpenFGA.EXPECT().ListUsers(gomock.Any(), "group#member", authz.MEMBER_RELATION, fmt.Sprintf("group:%s", test.group)).Return(test.expected.tuples.groups, test.expected.errGroupsListUsers)
			}

			if test.expected.errUsersListUsers == nil && test.expected.errGroupsListUsers != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			identities, err := svc.ListIdentities(context.Background(), test.group)

			if err != test.expected.errUsersListUsers && err != test.expected.errGroupsListUsers {
				t.Errorf("expected error to be one of %v and %v got %v", test.expected.errUsersListUsers, test.expected.errGroupsListUsers, err)
			}

			if test.expected.errUsersListUsers == nil && test.expected.errGroupsListUsers == nil && !reflect.DeepEqual(identities, test.output) {
				t.Errorf("invalid result, expected: %v, got: %v", test.output, identities)
			}
		})
	}
}

func TestServiceAssignRoles(t *testing.T) {
	type input struct {
		group string
		roles []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				group: "administrator",
				roles: []string{"viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple roles",
			input: input{
				group: "administrator",
				roles: []string{"viewer", "writer", "super"},
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.AssignRoles").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					roles := make([]ofga.Tuple, 0)

					for _, role := range test.input.roles {
						roles = append(roles, *ofga.NewTuple(fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION), authz.ASSIGNEE_RELATION, fmt.Sprintf("role:%s", role)))
					}

					if !reflect.DeepEqual(roles, tuples) {
						t.Errorf("expected tuples to be %v got %v", roles, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.AssignRoles(context.Background(), test.input.group, test.input.roles...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceCanAssignRoles(t *testing.T) {
	type input struct {
		roles []string
	}

	tests := []struct {
		name          string
		input         input
		expectedCheck bool
		expectedErr   error
	}{
		{
			name: "error",
			input: input{
				roles: []string{"joe"},
			},
			expectedCheck: false,
			expectedErr:   fmt.Errorf("error"),
		},
		{
			name: "multiple roles",
			input: input{
				roles: []string{"joe", "james", "ubork"},
			},
			expectedCheck: true,
			expectedErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			principalContext := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "mock-principal@email.com"})

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.CanAssignRoles").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().BatchCheck(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) (bool, error) {
					ids := make([]ofga.Tuple, 0)

					for _, r := range test.input.roles {
						ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:mock-principal@email.com"), authz.CAN_VIEW_RELATION, fmt.Sprintf("role:%s", r)))
					}

					if !reflect.DeepEqual(ids, tuples) {
						t.Errorf("expected tuples to be %v got %v", ids, tuples)
					}

					return test.expectedCheck, test.expectedErr
				},
			)

			if test.expectedErr != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			check, err := svc.CanAssignRoles(principalContext, "mock-principal@email.com", test.input.roles...)

			if err != test.expectedErr {
				t.Errorf("expected error to be %v got %v", test.expectedErr, err)
			}

			if check != test.expectedCheck {
				t.Errorf("expected check to be %v got %v", test.expectedCheck, err)
			}

		})
	}
}

func TestServiceRemoveRoles(t *testing.T) {
	type input struct {
		group string
		roles []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				group: "administrator",
				roles: []string{"viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple roles",
			input: input{
				group: "administrator",
				roles: []string{"viewer", "writer", "super"},
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.RemoveRoles").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					roles := make([]ofga.Tuple, 0)

					for _, role := range test.input.roles {
						roles = append(roles, *ofga.NewTuple(fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION), authz.ASSIGNEE_RELATION, fmt.Sprintf("role:%s", role)))
					}

					if !reflect.DeepEqual(roles, tuples) {
						t.Errorf("expected tuples to be %v got %v", roles, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.RemoveRoles(context.Background(), test.input.group, test.input.roles...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceAssignIdentities(t *testing.T) {
	type input struct {
		group      string
		identities []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				group:      "administrator",
				identities: []string{"joe"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple identities",
			input: input{
				group:      "administrator",
				identities: []string{"joe", "james", "ubork"},
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.AssignIdentities").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ids := make([]ofga.Tuple, 0)

					for _, i := range test.input.identities {
						ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:%s", i), authz.MEMBER_RELATION, fmt.Sprintf("group:%s", test.input.group)))
					}

					if !reflect.DeepEqual(ids, tuples) {
						t.Errorf("expected tuples to be %v got %v", ids, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.AssignIdentities(context.Background(), test.input.group, test.input.identities...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceCanAssignIdentities(t *testing.T) {
	type input struct {
		identities []string
	}

	tests := []struct {
		name          string
		input         input
		expectedCheck bool
		expectedErr   error
	}{
		{
			name: "error",
			input: input{
				identities: []string{"joe"},
			},
			expectedCheck: false,
			expectedErr:   fmt.Errorf("error"),
		},
		{
			name: "multiple identities",
			input: input{
				identities: []string{"joe", "james", "ubork"},
			},
			expectedCheck: true,
			expectedErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			principalContext := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "mock-principal@email.com"})

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.CanAssignIdentities").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().BatchCheck(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) (bool, error) {
					ids := make([]ofga.Tuple, 0)

					for _, i := range test.input.identities {
						ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:mock-principal@email.com"), authz.CAN_VIEW_RELATION, fmt.Sprintf("identity:%s", i)))
					}

					if !reflect.DeepEqual(ids, tuples) {
						t.Errorf("expected tuples to be %v got %v", ids, tuples)
					}

					return test.expectedCheck, test.expectedErr
				},
			)

			if test.expectedErr != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			check, err := svc.CanAssignIdentities(principalContext, "mock-principal@email.com", test.input.identities...)

			if err != test.expectedErr {
				t.Errorf("expected error to be %v got %v", test.expectedErr, err)
			}

			if check != test.expectedCheck {
				t.Errorf("expected check to be %v got %v", test.expectedCheck, err)
			}

		})
	}
}

func TestServiceRemoveIdentities(t *testing.T) {
	type input struct {
		group      string
		identities []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				group:      "administrator",
				identities: []string{"joe"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple identities",
			input: input{
				group:      "administrator",
				identities: []string{"joe", "james", "ubork"},
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.RemoveIdentities").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ids := make([]ofga.Tuple, 0)

					for _, i := range test.input.identities {
						ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:%s", i), authz.MEMBER_RELATION, fmt.Sprintf("group:%s", test.input.group)))
					}

					if !reflect.DeepEqual(ids, tuples) {
						t.Errorf("expected tuples to be %v got %v", ids, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := svc.RemoveIdentities(context.Background(), test.input.group, test.input.identities...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceGetGroup(t *testing.T) {
	type expected struct {
		err   error
		check bool
	}

	type input struct {
		group string
		user  string
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "not found",
			input: input{
				group: "administrator",
				user:  "admin",
			},
			expected: expected{
				check: false,
				err:   nil,
			},
		},
		{
			name: "error",
			input: input{
				group: "administrator",
				user:  "admin",
			},
			expected: expected{
				check: false,
				err:   fmt.Errorf("error"),
			},
		},
		{
			name: "found",
			input: input{
				group: "administrator",
				user:  "admin",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.GetGroup").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
				mockRepo.EXPECT().FindGroupByNameAndOwner(gomock.Any(), test.input.group, test.input.user).Times(1).Return(nil, test.expected.err)
			} else {
				mockRepo.EXPECT().FindGroupByNameAndOwner(gomock.Any(), test.input.group, test.input.user).Times(1).Return(&Group{ID: test.input.group, Name: test.input.group, Owner: test.input.user}, nil)
			}

			group, err := svc.GetGroup(context.Background(), test.input.user, test.input.group)

			if err != nil && err.Error() != fmt.Sprintf("unable to get group administrator for owner admin, %s", test.expected.err.Error()) {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && test.expected.check && group.ID != test.input.group {
				t.Errorf("invalid result, expected: %v, got: %v", test.input.group, group)
			}
		})
	}
}

func TestServiceCreateGroup(t *testing.T) {
	type input struct {
		group string
		user  string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				group: "administrator",
				user:  "admin",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "found",
			input: input{
				group: "administrator",
				user:  "admin",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)
			mockTx := NewMockTxInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.CreateGroup").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockRepo.EXPECT().CreateGroupTx(gomock.Any(), test.input.group, test.input.user).Times(1).Return(&Group{ID: test.input.group, Name: test.input.group, Owner: test.input.user}, mockTx, nil)

			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					ps = append(
						ps,
						*ofga.NewTuple(fmt.Sprintf("user:%s", test.input.user), authz.MEMBER_RELATION, fmt.Sprintf("group:%s", test.input.group)),
						*ofga.NewTuple(fmt.Sprintf("user:%s", test.input.user), authz.CAN_DELETE, fmt.Sprintf("group:%s", test.input.group)),
					)

					if !reflect.DeepEqual(ps, tuples) {
						t.Errorf("expected tuples to be %v got %v", ps, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
				mockTx.EXPECT().Rollback().Times(1).Return(nil)
			} else {
				mockTx.EXPECT().Commit().Times(1).Return(nil)
			}

			group, err := svc.CreateGroup(context.Background(), test.input.user, test.input.group)

			if test.expected != nil && err.Error() != test.expected.Error() {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}

			if group != nil && (group.ID != test.input.group || group.Name != test.input.group) {
				t.Errorf("expected group ID and Name to be %v got %s, %s", test.input.group, group.ID, group.Name)
			}
		})
	}
}

func TestServiceDeleteGroup(t *testing.T) {
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)
			for i := 0; i < 7; i++ {
				setupMockSubmit(workerPool, nil)
			}

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.DeleteGroup").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.removePermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.removeDirectAssociations").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			mockRepo.EXPECT().DeleteGroupByName(gomock.Any(), test.input).Times(1).Return(test.input, nil)

			pTypes := []string{"role", "group", "identity", "scheme", "provider", "client"}
			directRelations := []string{"privileged", "member", "can_create", "can_delete", "can_edit", "can_view"}

			calls := []*gomock.Call{}

			for _, pType := range pTypes {

				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), fmt.Sprintf("group:%s#%s", test.input, authz.MEMBER_RELATION), "", fmt.Sprintf("%s:", pType), "").Times(1).DoAndReturn(
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
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), "", relation, fmt.Sprintf("group:%s", test.input), "").Times(1).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										"user:test", authz.MEMBER_RELATION, object,
									),
									time.Now(),
								),
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										"group:test#member", authz.MEMBER_RELATION, object,
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

							if tuple.User != fmt.Sprintf("group:%s#%s", test.input, authz.MEMBER_RELATION) && tuple.User != authz.ADMIN_OBJECT {
								t.Errorf("expected user to be one of %v got %v", []string{fmt.Sprintf("group:%s#%s", test.input, authz.MEMBER_RELATION), authz.ADMIN_OBJECT}, tuple.User)
							}

							if tuple.Relation != "privileged" && tuple.Relation != "can_edit" {
								t.Errorf("expected relation to be one of %v got %v", []string{"privileged", "can_edit"}, tuple.Relation)
							}

							if tuple.Object != fmt.Sprintf("group:%s", test.input) && !strings.HasSuffix(tuple.Object, ":test") {
								t.Errorf("expected object to be one of %v got %v", []string{fmt.Sprintf("group:%s", test.input), "<*>:test"}, tuple.Object)
							}
						case 2:
							for _, tuple := range tuples {
								if tuple.User != "user:test" && tuple.User != "group:test#member" {
									t.Errorf("expected user to be one of %v got %v", []string{"user:test", "group:test#member"}, tuple.User)
								}

								if tuple.Relation != authz.MEMBER_RELATION {
									t.Errorf("expected relation to be of %v got %v", authz.MEMBER_RELATION, tuple.Relation)
								}

								if tuple.Object != fmt.Sprintf("group:%s", test.input) {
									t.Errorf("expected object to be one of %v got %v", fmt.Sprintf("group:%s", test.input), tuple.Object)
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

			_ = svc.DeleteGroup(context.Background(), test.input)

		})
	}
}

func TestServiceListPermissions(t *testing.T) {
	type input struct {
		group   string
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
				group: "administrator",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "found",
			input: input{
				group: "administrator",
				cTokens: map[string]string{
					"group": "test",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
			workerPool := NewMockWorkerPoolInterface(ctrl)
			for i := 0; i < 6; i++ {
				setupMockSubmit(workerPool, nil)
			}
			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.ListPermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.listPermissionsByType").Times(6).Return(context.TODO(), trace.SpanFromContext(context.TODO()))

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

							if user != fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION) {
								t.Errorf("wrong user parameter expected %s got %s", fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION), user)
							}

							if object == "group:" && continuationToken != "test" {
								t.Errorf("missing continuation token %s", test.input.cTokens["groups"])
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "can_edit", fmt.Sprintf("%stest", object),
									),
									time.Now(),
								),
							}

							if object == "role:role" {
								tuples = append(tuples,
									*openfga.NewTuple(
										*openfga.NewTupleKey(
											fmt.Sprintf("group:%s#%s", user, authz.MEMBER_RELATION), "assignee", fmt.Sprintf("%stest", object),
										),
										time.Now(),
									),
								)
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
				mockLogger.EXPECT().Errorf(gomock.Any()).AnyTimes()
			}

			gomock.InAnyOrder(calls)
			permissions, cTokens, err := svc.ListPermissions(context.Background(), test.input.group, test.input.cTokens)

			if err != nil && test.expected == nil {
				t.Errorf("expected error to be silenced and return nil got %v instead", err)
			}

			sort.Strings(permissions)
			sort.Strings(expPermissions)

			if err == nil && test.expected == nil && !reflect.DeepEqual(permissions, expPermissions) {
				t.Errorf("expected permissions to be %v got %v", expPermissions, permissions)
			}

			if err == nil && test.expected == nil && !reflect.DeepEqual(cTokens, expCTokens) {
				t.Errorf("expected continuation tokens to be %v got %v", expCTokens, cTokens)
			}
		})
	}
}

func TestServiceAssignPermissions(t *testing.T) {
	type input struct {
		group       string
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
				group: "administrator",
				permissions: []Permission{
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions",
			input: input{
				group: "administrator",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.AssignPermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *ofga.NewTuple(fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION), p.Relation, p.Object))
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

			err := svc.AssignPermissions(context.Background(), test.input.group, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestServiceRemovePermissions(t *testing.T) {
	type input struct {
		group       string
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
				group: "administrator",
				permissions: []Permission{
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions",
			input: input{
				group: "administrator",
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
			mockRepo := NewMockGroupRepositoryInterface(ctrl)

			workerPool := NewMockWorkerPoolInterface(ctrl)

			svc := NewService(mockOpenFGA, mockRepo, workerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), "groups.Service.RemovePermissions").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...ofga.Tuple) error {
					ps := make([]ofga.Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *ofga.NewTuple(fmt.Sprintf("group:%s#%s", test.input.group, authz.MEMBER_RELATION), p.Relation, p.Object))
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

			err := svc.RemovePermissions(context.Background(), test.input.group, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestV1Service_ListGroups(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		expectedResult []string
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "List groups successfully",
			setupMocks: func() {
				mockService.EXPECT().
					ListGroups(gomock.Any(), principal.Identifier()).
					Return([]string{"group1", "group2"}, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			expectedResult: []string{"group1", "group2"},
			expectedError:  nil,
		},
		{
			name:       "Unauthorized request",
			setupMocks: func() {},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectedResult: nil,
			expectedError:  v1.NewAuthorizationError("unauthorized"),
		},
		{
			name: "Error while listing groups",
			setupMocks: func() {
				mockService.EXPECT().
					ListGroups(gomock.Any(), principal.Identifier()).
					Return(nil, errors.New("some error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			expectedResult: nil,
			expectedError:  v1.NewUnknownError(fmt.Sprintf("failed to list groups for user %s: some error", principal.Identifier())),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.ListGroups(ctx, &resources.GetGroupsParams{})

			var groups []string
			if tc.expectedResult != nil {
				for _, group := range result.Data {
					groups = append(groups, group.Name)
				}
			}
			assert.Equal(t, tc.expectedResult, groups)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_CreateGroup(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		group          *resources.Group
		expectedResult string
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Create group successfully",
			setupMocks: func() {
				mockService.EXPECT().
					CreateGroup(gomock.Any(), principal.Identifier(), "group1").
					Return(&Group{ID: "1", Name: "group1"}, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			group:          &resources.Group{Name: "group1"},
			expectedResult: "group1",
			expectedError:  nil,
		},
		{
			name:       "Unauthorized request",
			setupMocks: func() {},
			contextSetup: func() context.Context {
				return context.Background()
			},
			group:         &resources.Group{Name: "group1"},
			expectedError: v1.NewAuthorizationError("unauthorized"),
		},
		{
			name: "Error while creating group",
			setupMocks: func() {
				mockService.EXPECT().
					CreateGroup(gomock.Any(), principal.Identifier(), "group1").
					Return(nil, errors.New("some error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			group:         &resources.Group{Name: "group1"},
			expectedError: v1.NewUnknownError(fmt.Sprintf("failed to create group group1 for user %s: some error", principal.Identifier())),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.CreateGroup(ctx, tc.group)

			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedResult, result.Name)
			}

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_GetGroup(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		expectedResult string
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Get group successfully",
			setupMocks: func() {
				mockService.EXPECT().
					GetGroup(gomock.Any(), principal.Identifier(), "group1").
					Return(&Group{ID: "1", Name: "group1"}, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			expectedResult: "group1",
			expectedError:  nil,
		},
		{
			name:       "Unauthorized request",
			setupMocks: func() {},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectedError: v1.NewAuthorizationError("unauthorized"),
		},
		{
			name: "Group not found",
			setupMocks: func() {
				mockService.EXPECT().
					GetGroup(gomock.Any(), principal.Identifier(), "group1").
					Return(nil, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			expectedError: v1.NewNotFoundError("group group1 not found"),
		},
		{
			name: "Error while getting group",
			setupMocks: func() {
				mockService.EXPECT().
					GetGroup(gomock.Any(), principal.Identifier(), "group1").
					Return(nil, errors.New("some error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			expectedError: v1.NewUnknownError(fmt.Sprintf("failed to get group group1 for user %s: some error", principal.Identifier())),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.GetGroup(ctx, "group1")

			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedResult, result.Name)
			}

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_UpdateGroup(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, _ := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		group          *resources.Group
		expectedResult *resources.Group
		expectedError  error
	}

	testCases := []testCase{
		{
			name:           "Not implemented",
			group:          &resources.Group{Name: "mock-group-name"},
			expectedResult: nil,
			expectedError:  v1.NewNotImplementedError("service not implemented"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.UpdateGroup(context.Background(), tc.group)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_DeleteGroup(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		groupId        string
		expectedResult bool
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Successfully deletes group",
			setupMocks: func() {
				mockService.EXPECT().DeleteGroup(gomock.Any(), "mock-group-id").Return(nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			groupId:        "mock-group-id",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:       "Unauthorized request",
			setupMocks: func() {},
			contextSetup: func() context.Context {
				return context.Background()
			},
			groupId:        "mock-group-id",
			expectedResult: false,
			expectedError:  v1.NewAuthorizationError("unauthorized"),
		},
		{
			name: "Error while deleting group",
			setupMocks: func() {
				mockService.EXPECT().DeleteGroup(gomock.Any(), "mock-group-id").Return(errors.New("some error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				ctx = authentication.PrincipalContext(ctx, principal)
				return ctx
			},
			groupId:        "mock-group-id",
			expectedResult: false,
			expectedError:  v1.NewUnknownError(fmt.Sprintf("failed to delete group mock-group-id for principal %s: some error", principal.Identifier())),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.DeleteGroup(ctx, tc.groupId)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_GetGroupIdentities(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	identities := []string{"identity1", "identity2"}

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		expectedResult *resources.PaginatedResponse[resources.Identity]
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Successfully retrieves group identities",
			setupMocks: func() {
				mockService.EXPECT().
					ListIdentities(gomock.Any(), "mock-group-id").
					Return([]string{"identity1", "identity2"}, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			expectedResult: &resources.PaginatedResponse[resources.Identity]{
				Meta: resources.ResponseMeta{Size: 2},
				Data: []resources.Identity{
					{Id: &identities[0]},
					{Id: &identities[1]},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error while retrieving group identities",
			setupMocks: func() {
				mockService.EXPECT().
					ListIdentities(gomock.Any(), "mock-group-id").
					Return(nil, errors.New("some error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			expectedResult: nil,
			expectedError:  v1.NewUnknownError("failed to list identities for group mock-group-id: some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.contextSetup()
			tc.setupMocks()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			_, err := s.GetGroupIdentities(ctx, "mock-group-id", &resources.GetGroupsItemIdentitiesParams{})

			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedResult.Meta.Size, len(identities))
			}
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_PatchGroupIdentities(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name            string
		setupMocks      func()
		contextSetup    func() context.Context
		identityPatches []resources.GroupIdentitiesPatchItem
		expectedResult  bool
		expectedError   error
	}

	testCases := []testCase{
		{
			name: "Successfully patches identities (add and remove)",
			setupMocks: func() {
				mockService.EXPECT().
					AssignIdentities(gomock.Any(), "mock-group-id", "identity1").
					Return(nil)

				mockService.EXPECT().
					RemoveIdentities(gomock.Any(), "mock-group-id", "identity2").
					Return(nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			identityPatches: []resources.GroupIdentitiesPatchItem{
				{Op: "add", Identity: "identity1"},
				{Op: "remove", Identity: "identity2"},
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "Error while assigning identities",
			setupMocks: func() {
				mockService.EXPECT().
					AssignIdentities(gomock.Any(), "mock-group-id", "identity1").
					Return(errors.New("assign error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			identityPatches: []resources.GroupIdentitiesPatchItem{
				{Op: "add", Identity: "identity1"},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to assign identities to group mock-group-id: assign error"),
		},
		{
			name: "Error while removing identities",
			setupMocks: func() {
				mockService.EXPECT().
					RemoveIdentities(gomock.Any(), "mock-group-id", "identity2").
					Return(errors.New("remove error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			identityPatches: []resources.GroupIdentitiesPatchItem{
				{Op: "remove", Identity: "identity2"},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to remove identities from group mock-group-id: remove error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.PatchGroupIdentities(ctx, "mock-group-id", tc.identityPatches)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_GetGroupRoles(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	roles := []string{"role1", "role2"}

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		params         *resources.GetGroupsItemRolesParams
		expectedResult []string
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Successfully retrieves roles",
			setupMocks: func() {
				mockService.EXPECT().
					ListRoles(gomock.Any(), "mock-group-id").
					Return(roles, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			params:         &resources.GetGroupsItemRolesParams{},
			expectedResult: roles,
			expectedError:  nil,
		},
		{
			name: "Error while retrieving roles",
			setupMocks: func() {
				mockService.EXPECT().
					ListRoles(gomock.Any(), "mock-group-id").
					Return(nil, errors.New("list roles error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			params:         &resources.GetGroupsItemRolesParams{},
			expectedResult: nil,
			expectedError:  v1.NewUnknownError("failed to list roles for group mock-group-id: list roles error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.GetGroupRoles(ctx, "mock-group-id", tc.params)

			var actualRoles []string
			if tc.expectedError == nil {
				for _, role := range result.Data {
					actualRoles = append(actualRoles, role.Name)
				}
			}

			assert.Equal(t, tc.expectedResult, actualRoles)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_PatchGroupRoles(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		rolePatches    []resources.GroupRolesPatchItem
		expectedResult bool
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Successfully patches roles (add and remove)",
			setupMocks: func() {
				mockService.EXPECT().
					AssignRoles(gomock.Any(), "mock-group-id", "role1").
					Return(nil)
				mockService.EXPECT().
					RemoveRoles(gomock.Any(), "mock-group-id", "role2").
					Return(nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			rolePatches: []resources.GroupRolesPatchItem{
				{Op: "add", Role: "role1"},
				{Op: "remove", Role: "role2"},
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "Error while assigning roles",
			setupMocks: func() {
				mockService.EXPECT().
					AssignRoles(gomock.Any(), "mock-group-id", "role1").
					Return(errors.New("assign roles error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			rolePatches: []resources.GroupRolesPatchItem{
				{Op: "add", Role: "role1"},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to assign roles to group mock-group-id: assign roles error"),
		},
		{
			name: "Error while removing roles",
			setupMocks: func() {
				mockService.EXPECT().
					AssignRoles(gomock.Any(), "mock-group-id", "role1").
					Return(nil)
				mockService.EXPECT().
					RemoveRoles(gomock.Any(), "mock-group-id", "role2").
					Return(errors.New("remove roles error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			rolePatches: []resources.GroupRolesPatchItem{
				{Op: "add", Role: "role1"},
				{Op: "remove", Role: "role2"},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to remove roles from group mock-group-id: remove roles error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.PatchGroupRoles(ctx, "mock-group-id", tc.rolePatches)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_GetGroupEntitlements(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	permissions := []string{"can_view::client:okta", "can_view::group:admin"}
	currPageToken := map[string]string{
		"groups": "page-token",
	}
	nextPageToken := map[string]string{
		"groups": "new-page-token",
	}

	paginator := types.NewTokenPaginator(mockTracer, mockLogger)

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		groupId        string
		expectedResult *resources.PaginatedResponse[resources.EntityEntitlement]
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "Successfully retrieves group entitlements",
			setupMocks: func() {
				mockService.EXPECT().
					ListPermissions(gomock.Any(), "mock-group-id", currPageToken).
					Return(permissions, nextPageToken, nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			expectedResult: &resources.PaginatedResponse[resources.EntityEntitlement]{
				Meta: resources.ResponseMeta{Size: 2},
				Data: []resources.EntityEntitlement{
					{Entitlement: "can_view", EntityType: "client", EntityId: "okta"},
					{Entitlement: "can_view", EntityType: "group", EntityId: "admin"},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error while retrieving permissions",
			setupMocks: func() {
				mockService.EXPECT().
					ListPermissions(gomock.Any(), "mock-group-id", currPageToken).
					Return(nil, nil, errors.New("permissions error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			expectedResult: nil,
			expectedError:  v1.NewUnknownError("failed to list permissions for group mock-group-id: permissions error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()
			paginator.SetTokens(ctx, currPageToken)
			pageToken, _ := paginator.PaginationHeader(ctx)

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.GetGroupEntitlements(ctx, "mock-group-id", &resources.GetGroupsItemEntitlementsParams{NextToken: &pageToken})

			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedResult.Meta, result.Meta)
				assert.Equal(t, tc.expectedResult.Data, result.Data)

				paginator.SetTokens(ctx, nextPageToken)
				expectedToken, _ := paginator.PaginationHeader(ctx)
				assert.Equal(t, expectedToken, *result.Next.PageToken)

			}
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_PatchGroupEntitlements(t *testing.T) {
	ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name               string
		setupMocks         func()
		contextSetup       func() context.Context
		entitlementPatches []resources.GroupEntitlementsPatchItem
		expectedResult     bool
		expectedError      error
	}

	testCases := []testCase{
		{
			name: "Successfully patches entitlements - add and remove",
			setupMocks: func() {
				mockService.EXPECT().
					AssignPermissions(gomock.Any(), "mock-group-id", gomock.Any()).
					Return(nil)
				mockService.EXPECT().
					RemovePermissions(gomock.Any(), "mock-group-id", gomock.Any()).
					Return(nil)
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			entitlementPatches: []resources.GroupEntitlementsPatchItem{
				{Op: "add", Entitlement: resources.EntityEntitlement{Entitlement: "can_view", EntityType: "client", EntityId: "okta"}},
				{Op: "remove", Entitlement: resources.EntityEntitlement{Entitlement: "can_view", EntityType: "group", EntityId: "admin"}},
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "Error while assigning permissions",
			setupMocks: func() {
				mockService.EXPECT().
					AssignPermissions(gomock.Any(), "mock-group-id", gomock.Any()).
					Return(errors.New("assignment error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			entitlementPatches: []resources.GroupEntitlementsPatchItem{
				{Op: "add", Entitlement: resources.EntityEntitlement{Entitlement: "can_view", EntityType: "client", EntityId: "okta"}},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to assign permissions to group mock-group-id: assignment error"),
		},
		{
			name: "Error while removing permissions",
			setupMocks: func() {
				mockService.EXPECT().
					AssignPermissions(gomock.Any(), "mock-group-id", gomock.Any()).
					Return(nil)
				mockService.EXPECT().
					RemovePermissions(gomock.Any(), "mock-group-id", gomock.Any()).
					Return(errors.New("removal error"))
			},
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			entitlementPatches: []resources.GroupEntitlementsPatchItem{
				{Op: "add", Entitlement: resources.EntityEntitlement{Entitlement: "can_view", EntityType: "client", EntityId: "okta"}},
				{Op: "remove", Entitlement: resources.EntityEntitlement{Entitlement: "can_view", EntityType: "group", EntityId: "admin"}},
			},
			expectedResult: false,
			expectedError:  v1.NewUnknownError("failed to remove permissions from group mock-group-id: removal error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(mockService, mockTracer, mockMonitor, mockLogger)

			result, err := s.PatchGroupEntitlements(ctx, "mock-group-id", tc.entitlementPatches)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func setupTest(t *testing.T) (
	*gomock.Controller,
	*MockServiceInterface,
	*MockLoggerInterface,
	*MockTracer,
	*monitoring.MockMonitorInterface,
	*authentication.ServicePrincipal,
) {
	ctrl := gomock.NewController(t)
	mockService := NewMockServiceInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockProvider := NewMockProviderInterface(ctrl)

	mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
			return ctx, trace.SpanFromContext(ctx)
		},
	)
	mockProvider.EXPECT().Verifier(gomock.Any()).Return(
		oidc.NewVerifier("", nil, &oidc.Config{
			ClientID:                   "mock-client-id",
			SkipExpiryCheck:            true,
			SkipIssuerCheck:            true,
			InsecureSkipSignatureCheck: true,
		}),
	)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJtb2NrLWNsaWVudC1pZCIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.BdspASNsnxeXnqZXZnFnkvv-ClMq0U6X1gCIUrh9V7c"
	principal, _ := authentication.NewJWKSTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor).VerifyAccessToken(context.TODO(), token)

	return ctrl, mockService, mockLogger, mockTracer, mockMonitor, principal
}
