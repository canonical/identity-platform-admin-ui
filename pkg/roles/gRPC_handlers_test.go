// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	v0Roles "github.com/canonical/identity-platform-api/v0/roles"

	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_validation.go -source=../../internal/validation/registry.go

func strPtr(s string) *string {
	return &s
}

func TestListRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		expected *v0Roles.ListRolesResp
		err      error
	}{
		{
			name: "Successful list",
			expected: &v0Roles.ListRolesResp{
				Data:    []string{"admin", "user"},
				Status:  http.StatusOK,
				Message: strPtr("List of roles"),
			},
			err: nil,
		},
		{
			name:     "Service error",
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list roles"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.ListRoles").Return(mockCtx, mockSpan)
			if test.err != nil {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var expectedData []string = nil
			if test.expected != nil {
				expectedData = test.expected.Data
			}
			mockService.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(expectedData, test.err)

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.ListRoles(mockCtx, &emptypb.Empty{})

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if !reflect.DeepEqual(resp.Data, test.expected.Data) {
					t.Errorf("expected data: %v, got: %v", test.expected.Data, resp.Data)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.GetRoleReq
		expected *v0Roles.GetRoleResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Roles.GetRoleReq{Id: "role1"},
			expected: &v0Roles.GetRoleResp{
				Data:    []*v0Roles.Role{{Id: "role1", Name: "role1"}},
				Status:  http.StatusOK,
				Message: strPtr("Role detail"),
			},
			err: nil,
		},
		{
			name:     "Role not found",
			req:      &v0Roles.GetRoleReq{Id: "unknown"},
			expected: nil,
			err:      status.Errorf(codes.NotFound, "role not found: unknown"),
		},
		{
			name:     "Service error",
			req:      &v0Roles.GetRoleReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to retrieve role: role1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.GetRole").Return(mockCtx, mockSpan)

			if test.name == "Role not found" {
				mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			}

			if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var expectedRole *Role = nil
			if test.expected != nil {
				expectedRole = &Role{ID: test.expected.Data[0].Id, Name: test.expected.Data[0].Name}
			}

			if test.name == "Role not found" {
				mockService.EXPECT().GetRole(gomock.Any(), gomock.Any(), test.req.GetId()).Return(nil, nil)
			} else {
				mockService.EXPECT().GetRole(gomock.Any(), gomock.Any(), test.req.GetId()).Return(expectedRole, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.GetRole(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if !reflect.DeepEqual(resp.Data, test.expected.Data) {
					t.Errorf("expected data: %v, got: %v", test.expected.Data, resp.Data)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestCreateRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.CreateRoleReq
		expected *v0Roles.CreateRoleResp
		err      error
	}{
		{
			name: "Successful creation",
			req:  &v0Roles.CreateRoleReq{Role: &v0Roles.Role{Name: "Admin"}},
			expected: &v0Roles.CreateRoleResp{
				Data:    []*v0Roles.Role{{Id: "role1", Name: "Admin"}},
				Status:  http.StatusCreated,
				Message: strPtr("Created role Admin"),
			},
			err: nil,
		},
		{
			name:     "Empty role name",
			req:      &v0Roles.CreateRoleReq{Role: &v0Roles.Role{Name: ""}},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role name is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Roles.CreateRoleReq{Role: &v0Roles.Role{Name: "Admin"}},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to create role: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.CreateRole").Return(mockCtx, mockSpan)

			if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var createdRole *Role = nil
			if test.expected != nil {
				createdRole = &Role{ID: "role1", Name: test.expected.Data[0].Name}
			}

			if test.name != "Empty role name" {
				mockService.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any()).Return(createdRole, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.CreateRole(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if !reflect.DeepEqual(resp.Data, test.expected.Data) {
					t.Errorf("expected data: %v, got: %v", test.expected.Data, resp.Data)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestRemoveRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.RemoveRoleReq
		expected *v0Roles.RemoveRoleResp
		err      error
	}{
		{
			name: "Successful removal",
			req:  &v0Roles.RemoveRoleReq{Id: "role1"},
			expected: &v0Roles.RemoveRoleResp{
				Status:  http.StatusOK,
				Message: strPtr("Deleted role role1"),
			},
			err: nil,
		},
		{
			name:     "Empty role ID",
			req:      &v0Roles.RemoveRoleReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Roles.RemoveRoleReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to delete role: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.RemoveRole").Return(mockCtx, mockSpan)

			if test.name == "Empty role ID" {
				mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty role ID" {
				mockService.EXPECT().DeleteRole(gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveRole(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestListRoleEntitlements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.ListRoleEntitlementsReq
		expected *v0Roles.ListRoleEntitlementsResp
		err      error
	}{
		{
			name: "Successful list",
			req:  &v0Roles.ListRoleEntitlementsReq{Id: "role1"},
			expected: &v0Roles.ListRoleEntitlementsResp{
				Data:    []string{"can_view::identity:id1"},
				Status:  http.StatusOK,
				Message: strPtr("List of entitlements"),
			},
			err: nil,
		},
		{
			name:     "Empty role ID",
			req:      &v0Roles.ListRoleEntitlementsReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name:     "Paginator load error",
			req:      &v0Roles.ListRoleEntitlementsReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to load token paginator: some error"),
		},
		{
			name:     "Service error",
			req:      &v0Roles.ListRoleEntitlementsReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list entitlements for role role1, some error"),
		},
		{
			name:     "Pagination metadata error",
			req:      &v0Roles.ListRoleEntitlementsReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "error producing pagination metadata: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(mockCtx, mockSpan)

			if test.name == "Empty role ID" {
				mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			} else if test.name == "Paginator load error" || test.name == "Pagination metadata error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
			}

			var serviceReturnValues []string
			if test.expected != nil {
				serviceReturnValues = test.expected.Data
			}

			if test.name != "Empty role ID" {
				mockService.EXPECT().ListPermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(serviceReturnValues, nil, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.ListRoleEntitlements(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if !reflect.DeepEqual(resp.Data, test.expected.Data) {
					t.Errorf("expected data: %v, got: %v", test.expected.Data, resp.Data)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestUpdateRoleEntitlements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.UpdateRoleEntitlementsReq
		expected *v0Roles.UpdateRoleEntitlementsResp
		err      error
	}{
		{
			name: "Successful update",
			req: &v0Roles.UpdateRoleEntitlementsReq{
				Id: "role1",
				EntitlementsPatchReq: &v0Roles.Permissions{
					Updates: []*v0Roles.Permission{
						{Relation: "relation1", Object: "object1"},
					},
				},
			},
			expected: &v0Roles.UpdateRoleEntitlementsResp{
				Status:  http.StatusOK,
				Message: strPtr("Updated permissions for role role1"),
			},
			err: nil,
		},
		{
			name: "Empty role ID",
			req: &v0Roles.UpdateRoleEntitlementsReq{
				Id: "",
				EntitlementsPatchReq: &v0Roles.Permissions{
					Updates: []*v0Roles.Permission{
						{Relation: "relation1", Object: "object1"},
					},
				},
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Roles.UpdateRoleEntitlementsReq{
				Id: "role1",
				EntitlementsPatchReq: &v0Roles.Permissions{
					Updates: []*v0Roles.Permission{
						{Relation: "relation1", Object: "object1"},
					},
				},
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to assign permissions: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.UpdateRoleEntitlements").Return(mockCtx, mockSpan)

			if test.name == "Empty role ID" {
				mockLogger.EXPECT().Debugf(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty role ID" {
				mockService.EXPECT().AssignPermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.UpdateRoleEntitlements(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if resp.Status != test.expected.Status {
					t.Errorf("expected status: %v, got: %v", test.expected.Status, resp.Status)
				}
				if resp.Message == nil || *resp.Message != *test.expected.Message {
					t.Errorf("expected message: %v, got: %v", *test.expected.Message, resp.Message)
				}
			}
		})
	}
}

func TestRemoveRoleEntitlement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.RemoveRoleEntitlementReq
		expected *v0Roles.RemoveRoleEntitlementResp
		err      error
	}{
		{
			name: "Successful removal",
			req: &v0Roles.RemoveRoleEntitlementReq{
				Id:            "role1",
				EntitlementId: "can_edit::client:okta",
			},
			expected: &v0Roles.RemoveRoleEntitlementResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed permission can_edit for role role1"),
			},
			err: nil,
		},
		{
			name: "Empty role ID",
			req: &v0Roles.RemoveRoleEntitlementReq{
				Id:            "",
				EntitlementId: "can_edit::client:okta",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name: "Empty entitlement ID",
			req: &v0Roles.RemoveRoleEntitlementReq{
				Id:            "role1",
				EntitlementId: "",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "entitlement ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Roles.RemoveRoleEntitlementReq{
				Id:            "role1",
				EntitlementId: "can_edit::client:okta",
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to remove permission relation1 from role role1: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.RemoveRoleEntitlement").Return(mockCtx, mockSpan)

			if test.name == "Empty role ID" || test.name == "Empty entitlement ID" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty role ID" && test.name != "Empty entitlement ID" {
				mockService.EXPECT().RemovePermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveRoleEntitlement(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if !reflect.DeepEqual(resp, test.expected) {
					t.Errorf("expected: %v, got: %v", test.expected, resp)
				}
			}
		})
	}
}

func TestGetRoleGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Roles.GetRoleGroupsReq
		expected *v0Roles.GetRoleGroupsResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Roles.GetRoleGroupsReq{Id: "role1"},
			expected: &v0Roles.GetRoleGroupsResp{
				Data:    []string{"group1"},
				Status:  http.StatusOK,
				Message: strPtr("List of groups"),
			},
			err: nil,
		},
		{
			name:     "Empty role ID",
			req:      &v0Roles.GetRoleGroupsReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Roles.GetRoleGroupsReq{Id: "role1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list role ID role1 groups: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})

			mockTracer.EXPECT().Start(gomock.Any(), "roles.GrpcHandler.GetRoleGroups").Return(mockCtx, mockSpan)

			if test.name == "Empty role ID" {
				mockLogger.EXPECT().Debugf(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty role ID" {
				var expectedData []string = nil
				if test.expected != nil {
					expectedData = test.expected.Data
				}
				mockService.EXPECT().ListRoleGroups(gomock.Any(), gomock.Any()).Return(expectedData, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.GetRoleGroups(mockCtx, test.req)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if err == nil {
				if !reflect.DeepEqual(resp, test.expected) {
					t.Errorf("expected: %v, got: %v", test.expected, resp)
				}
			}
		})
	}
}
