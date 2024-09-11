// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package openfga

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	sync "sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	pool "github.com/canonical/identity-platform-admin-ui/internal/pool"
)

//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_client.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_openfga_client.go github.com/openfga/go-sdk/client SdkClientListObjectsRequestInterface,SdkClientReadRequestInterface,SdkClientWriteRequestInterface,SdkClientBatchCheckRequestInterface
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_pool.go -source=../../internal/pool/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

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

func TestStoreListViewableRoles(t *testing.T) {
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
			input: "user:joe",
			expected: expected{
				roles: []string{},
				err:   nil,
			},
		},
		{
			name:  "error",
			input: "role:administrator#assignee",
			expected: expected{
				roles: []string{},
				err:   fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: "group:is#member",
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ListObjects(gomock.Any(), test.input, "can_view", "role").Return(test.expected.roles, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			roles, err := store.ListViewableRoles(context.Background(), test.input)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(roles, test.expected.roles) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.roles, roles)
			}
		})
	}
}

func TestStoreListAssignedRoles(t *testing.T) {
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
			input: "user:joe",
			expected: expected{
				roles: []string{},
				err:   nil,
			},
		},
		{
			name:  "error",
			input: "user:joe",
			expected: expected{
				roles: []string{},
				err:   fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: "group:is#member",
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ListObjects(gomock.Any(), test.input, ASSIGNEE_RELATION, "role").Return(test.expected.roles, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			roles, err := store.ListAssignedRoles(context.Background(), test.input)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(roles, test.expected.roles) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.roles, roles)
			}
		})
	}
}

func TestStoreListAssignedGroups(t *testing.T) {
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
			input: "user:joe",
			expected: expected{
				groups: []string{},
				err:    nil,
			},
		},
		{
			name:  "error",
			input: "user:joe",
			expected: expected{
				groups: []string{},
				err:    fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: "group:is#member",
			expected: expected{
				groups: []string{"global", "administrator", "viewer"},
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().ListObjects(gomock.Any(), test.input, MEMBER_RELATION, "group").Return(test.expected.groups, test.expected.err)

			if test.expected.err != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			groups, err := store.ListAssignedGroups(context.Background(), test.input)

			if err != test.expected.err {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err == nil && !reflect.DeepEqual(groups, test.expected.groups) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.groups, groups)
			}
		})
	}
}

func TestStoreAssignRoles(t *testing.T) {
	type input struct {
		assignee string
		roles    []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				assignee: "group:administrator#member",
				roles:    []string{"role:viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple roles to group members",
			input: input{
				assignee: "group:administrator#member",
				roles:    []string{"role:viewer", "role:writer", "role:super"},
			},
			expected: nil,
		},
		{
			name: "multiple roles to a user",
			input: input{
				assignee: "user:joe",
				roles:    []string{"role:viewer", "role:writer", "role:super"},
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					roles := make([]Tuple, 0)

					for _, role := range test.input.roles {
						roles = append(roles, *NewTuple(test.input.assignee, ASSIGNEE_RELATION, role))
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

			err := store.AssignRoles(context.Background(), test.input.assignee, test.input.roles...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreUnassignRoles(t *testing.T) {
	type input struct {
		assignee string
		roles    []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				assignee: "group:administrator#member",
				roles:    []string{"role:viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple roles to group members",
			input: input{
				assignee: "group:administrator#member",
				roles:    []string{"role:viewer", "role:writer", "role:super"},
			},
			expected: nil,
		},
		{
			name: "multiple roles to a user",
			input: input{
				assignee: "user:joe",
				roles:    []string{"role:viewer", "role:writer", "role:super"},
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					roles := make([]Tuple, 0)

					for _, role := range test.input.roles {
						roles = append(roles, *NewTuple(test.input.assignee, ASSIGNEE_RELATION, role))
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

			err := store.UnassignRoles(context.Background(), test.input.assignee, test.input.roles...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreAssignGroups(t *testing.T) {
	type input struct {
		assignee string
		groups   []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				assignee: "administrator",
				groups:   []string{"group:viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple groups to group members",
			input: input{
				assignee: "group:administrator#member",
				groups:   []string{"group:viewer", "group:writer", "group:super"},
			},
			expected: nil,
		},
		{
			name: "multiple groups to a user",
			input: input{
				assignee: "user:joe",
				groups:   []string{"group:viewer", "group:writer", "group:super"},
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					groups := make([]Tuple, 0)

					for _, group := range test.input.groups {
						groups = append(groups, *NewTuple(test.input.assignee, MEMBER_RELATION, group))
					}

					if !reflect.DeepEqual(groups, tuples) {
						t.Errorf("expected tuples to be %v got %v", groups, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := store.AssignGroups(context.Background(), test.input.assignee, test.input.groups...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreUnassignGroups(t *testing.T) {
	type input struct {
		assignee string
		groups   []string
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				assignee: "administrator",
				groups:   []string{"group:viewer"},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple groups to group members",
			input: input{
				assignee: "group:administrator#member",
				groups:   []string{"group:viewer", "group:writer", "group:super"},
			},
			expected: nil,
		},
		{
			name: "multiple groups to a user",
			input: input{
				assignee: "user:joe",
				groups:   []string{"group:viewer", "group:writer", "group:super"},
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					groups := make([]Tuple, 0)

					for _, group := range test.input.groups {
						groups = append(groups, *NewTuple(test.input.assignee, MEMBER_RELATION, group))
					}

					if !reflect.DeepEqual(groups, tuples) {
						t.Errorf("expected tuples to be %v got %v", groups, tuples)
					}

					return test.expected
				},
			)

			if test.expected != nil {
				mockLogger.EXPECT().Error(gomock.Any()).Times(1)
			}

			err := store.UnassignGroups(context.Background(), test.input.assignee, test.input.groups...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreAssignPermissions(t *testing.T) {
	type input struct {
		assignee    string
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
				assignee: "role:administrator#assignee",
				permissions: []Permission{
					{Relation: "can_delete", Object: "role:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions to role",
			input: input{
				assignee: "role:administrator#assignee",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: nil,
		},
		{
			name: "multiple permissions to group",
			input: input{
				assignee: "group:administrator#member",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: nil,
		},
		{
			name: "multiple permissions to user",
			input: input{
				assignee: "user:joe",
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().WriteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					ps := make([]Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *NewTuple(test.input.assignee, p.Relation, p.Object))
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

			err := store.AssignPermissions(context.Background(), test.input.assignee, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreUnassignPermissions(t *testing.T) {
	type input struct {
		assignee    string
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
				assignee: "role:administrator#assignee",
				permissions: []Permission{
					{Relation: "can_delete", Object: "role:admin"},
				},
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "multiple permissions to role",
			input: input{
				assignee: "role:administrator#assignee",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: nil,
		},
		{
			name: "multiple permissions to group",
			input: input{
				assignee: "group:administrator#member",
				permissions: []Permission{
					{Relation: "can_view", Object: "client:okta"},
					{Relation: "can_edit", Object: "client:okta"},
					{Relation: "can_delete", Object: "group:admin"},
				},
			},
			expected: nil,
		},
		{
			name: "multiple permissions to user",
			input: input{
				assignee: "user:joe",
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGA.EXPECT().DeleteTuples(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, tuples ...Tuple) error {
					ps := make([]Tuple, 0)

					for _, p := range test.input.permissions {
						ps = append(ps, *NewTuple(test.input.assignee, p.Relation, p.Object))
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

			err := store.UnassignPermissions(context.Background(), test.input.assignee, test.input.permissions...)

			if err != test.expected {
				t.Errorf("expected error to be %v got %v", test.expected, err)
			}
		})
	}
}

func TestStoreListPermissions(t *testing.T) {
	type input struct {
		ID      string
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
				ID: "role:administrator#assignee",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "role found",
			input: input{
				ID: "role:administrator#assignee",
				cTokens: map[string]string{
					"role": "test",
				},
			},
			expected: nil,
		},
		{
			name: "group found",
			input: input{
				ID: "group:administrator#member",
				cTokens: map[string]string{
					"role": "test",
				},
			},
			expected: nil,
		},
		{
			name: "user found",
			input: input{
				ID: "use:joe",
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			for i := 0; i < 6; i++ {
				setupMockSubmit(mockWorkerPool, nil)
			}

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))

			expCTokens := map[string]string{
				"role":     "",
				"group":    "",
				"identity": "",
				"scheme":   "",
				"provider": "",
				"client":   "",
			}

			expPermissions := []Permission{
				{Relation: "can_edit", Object: "role:test"},
				{Relation: "can_edit", Object: "group:test"},
				{Relation: "can_edit", Object: "identity:test"},
				{Relation: "can_edit", Object: "scheme:test"},
				{Relation: "can_edit", Object: "provider:test"},
				{Relation: "can_edit", Object: "client:test"},
			}

			calls := []*gomock.Call{}

			for _, _ = range store.permissionTypes() {
				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							if user != test.input.ID {
								t.Errorf("wrong user parameter expected %s got %s", test.input.ID, user)
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
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "assignee", "role:test",
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
			permissions, cTokens, err := store.ListPermissions(context.Background(), test.input.ID, test.input.cTokens)

			if err != nil && test.expected == nil {
				t.Fatalf("expected error to be silenced and return nil got %v instead", err)
			}

			sortFx := func(a, b Permission) int {
				if n := strings.Compare(a.Relation, b.Relation); n != 0 {
					return n
				}
				// If relations are equal, order by object
				return cmp.Compare(a.Object, b.Object)
			}

			slices.SortFunc(permissions, sortFx)
			slices.SortFunc(expPermissions, sortFx)

			if err == nil && test.expected == nil && !reflect.DeepEqual(permissions, expPermissions) {
				t.Fatalf("expected permissions to be %v got %v", expPermissions, permissions)
			}

			if err == nil && test.expected == nil && !reflect.DeepEqual(cTokens, expCTokens) {
				t.Fatalf("expected continuation tokens to be %v got %v", expCTokens, cTokens)
			}
		})
	}
}

func TestStoreListPermissionsWithPermissions(t *testing.T) {
	type input struct {
		ID             string
		relationFilter *RelationFilter
		typesFilter    *TypesFilter
		tokenMapFilter *TokenMapFilter
	}

	tests := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "error",
			input: input{
				ID: "role:administrator#assignee",
			},
			expected: fmt.Errorf("error"),
		},
		{
			name: "role found",
			input: input{
				ID:             "role:administrator#assignee",
				relationFilter: NewRelationFilter("can_edit"),
				tokenMapFilter: NewTokenMapFilter(
					map[string]string{"role": "test"},
				),
			},
			expected: nil,
		},
		{
			name: "group found",
			input: input{
				ID:          "group:administrator#member",
				typesFilter: NewTypesFilter("identity", "client"),
				tokenMapFilter: NewTokenMapFilter(
					map[string]string{"role": "test"},
				),
			},
			expected: nil,
		},
		{
			name: "user found",
			input: input{
				ID: "use:joe",
				tokenMapFilter: NewTokenMapFilter(
					map[string]string{"role": "test"},
				),
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
			mockWorkerPool := NewMockWorkerPoolInterface(ctrl)

			store := NewOpenFGAStore(mockOpenFGA, mockWorkerPool, mockTracer, mockMonitor, mockLogger)

			types := store.permissionTypes()

			if test.input.typesFilter != nil {
				types = test.input.typesFilter.WithFilter().([]string)
			}

			for i := 0; i < len(types); i++ {
				setupMockSubmit(mockWorkerPool, nil)
			}

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Errorf(gomock.Any()).AnyTimes()

			expCTokens := make(map[string]string)
			expPermissions := make([]Permission, 0)

			for _, t := range types {
				expPermissions = append(
					expPermissions,
					Permission{Relation: "can_edit", Object: t + ":test"},
				)
				expCTokens[t] = ""
			}

			calls := []*gomock.Call{}

			for _, _ = range types {
				calls = append(
					calls,
					mockOpenFGA.EXPECT().ReadTuples(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
						func(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
							if test.expected != nil {
								return nil, test.expected
							}

							if user != test.input.ID {
								t.Errorf("wrong user parameter expected %s got %s", test.input.ID, user)
							}

							if object == "role:" && continuationToken != "test" {
								tokenM, ok := test.input.tokenMapFilter.WithFilter().(map[string]string)

								if !ok {
									t.Fatal("failed parsing token map")
								}

								t.Errorf("missing continuation token %s", tokenM["roles"])
							}

							tuples := []openfga.Tuple{
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "can_edit", fmt.Sprintf("%stest", object),
									),
									time.Now(),
								),
								*openfga.NewTuple(
									*openfga.NewTupleKey(
										user, "assignee", "role:test",
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

			gomock.InAnyOrder(calls)
			permissions, cTokens, err := store.ListPermissionsWithFilters(context.Background(), test.input.ID, test.input.typesFilter, test.input.tokenMapFilter, test.input.relationFilter)

			if err != nil && test.expected == nil {
				t.Fatalf("expected error to be silenced and return nil got %v instead", err)
			}

			sortFx := func(a, b Permission) int {
				if n := strings.Compare(a.Relation, b.Relation); n != 0 {
					return n
				}
				// If relations are equal, order by object
				return cmp.Compare(a.Object, b.Object)
			}

			slices.SortFunc(permissions, sortFx)
			slices.SortFunc(expPermissions, sortFx)

			if err == nil && test.expected == nil && !reflect.DeepEqual(permissions, expPermissions) {
				t.Fatalf("expected permissions to be %v got %v", expPermissions, permissions)
			}

			if err == nil && test.expected == nil && !reflect.DeepEqual(cTokens, expCTokens) {
				t.Fatalf("expected continuation tokens to be %v got %v", expCTokens, cTokens)
			}
		})
	}
}
