// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	v0Types "github.com/canonical/identity-platform-api/v0/http"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	v0Groups "github.com/canonical/identity-platform-api/v0/groups"

	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func strPtr(s string) *string {
	return &s
}

func TestListGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		expected *v0Groups.ListGroupsResp
		err      error
	}{
		{
			name: "Successful list",
			expected: &v0Groups.ListGroupsResp{
				Data:    []string{"group1", "group2"},
				Status:  http.StatusOK,
				Message: strPtr("List of groups"),
			},
			err: nil,
		},
		{
			name:     "Service error",
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list groups"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.ListGroups").Return(mockCtx, mockSpan)
			if test.err != nil {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var expectedData []string = nil
			if test.expected != nil {
				expectedData = test.expected.Data
			}
			mockService.EXPECT().ListGroups(gomock.Any(), gomock.Any()).Return(expectedData, test.err)

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.ListGroups(mockCtx, &emptypb.Empty{})

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

func TestGetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.GetGroupReq
		expected *v0Groups.GetGroupResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Groups.GetGroupReq{Id: "group1"},
			expected: &v0Groups.GetGroupResp{
				Data:    []*v0Groups.Group{{Id: "group1", Name: "group1"}},
				Status:  http.StatusOK,
				Message: strPtr("Group detail"),
			},
			err: nil,
		},
		{
			name:     "Group not found",
			req:      &v0Groups.GetGroupReq{Id: "unknown"},
			expected: nil,
			err:      status.Errorf(codes.NotFound, "group not found: unknown"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.GetGroupReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to retrieve group: group1"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.GetGroup").Return(mockCtx, mockSpan)

			if test.name == "Group not found" {
				mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			}

			if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var expectedRole *Group = nil
			if test.expected != nil {
				expectedRole = &Group{ID: test.expected.Data[0].Id, Name: test.expected.Data[0].Name}
			}

			if test.name == "Group not found" {
				mockService.EXPECT().GetGroup(gomock.Any(), gomock.Any(), test.req.GetId()).Return(nil, nil)
			} else {
				mockService.EXPECT().GetGroup(gomock.Any(), gomock.Any(), test.req.GetId()).Return(expectedRole, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.GetGroup(mockCtx, test.req)

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

func TestCreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.CreateGroupReq
		expected *v0Groups.CreateGroupResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Groups.CreateGroupReq{Group: &v0Groups.Group{Name: "group1"}},
			expected: &v0Groups.CreateGroupResp{
				Data:    []*v0Groups.Group{{Id: "group1", Name: "group1"}},
				Status:  http.StatusCreated,
				Message: strPtr("Created group group1"),
			},
			err: nil,
		},
		{
			name:     "Empty group name",
			req:      &v0Groups.CreateGroupReq{Group: &v0Groups.Group{Name: ""}},
			expected: nil,
			err:      status.Errorf(codes.NotFound, "group name is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.CreateGroupReq{Group: &v0Groups.Group{Name: "group1"}},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to create group: some-error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.CreateGroup").Return(mockCtx, mockSpan)

			if test.name == "Empty group name" {
				mockLogger.EXPECT().Debug(gomock.Any())
			}

			if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			var createdRole *Group = nil
			if test.expected != nil {
				createdRole = &Group{ID: test.expected.Data[0].Name, Name: test.expected.Data[0].Name}
			}

			if test.name != "Empty group name" {
				mockService.EXPECT().CreateGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(createdRole, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.CreateGroup(mockCtx, test.req)

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

func TestRemoveGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.RemoveGroupReq
		expected *v0Groups.RemoveGroupResp
		err      error
	}{
		{
			name: "Successful removal",
			req:  &v0Groups.RemoveGroupReq{Id: "group1"},
			expected: &v0Groups.RemoveGroupResp{
				Status:  http.StatusOK,
				Message: strPtr("Deleted group group1"),
			},
			err: nil,
		},
		{
			name:     "Empty group ID",
			req:      &v0Groups.RemoveGroupReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.RemoveGroupReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to delete group: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.RemoveGroup").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				mockService.EXPECT().DeleteGroup(gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveGroup(mockCtx, test.req)

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

func TestListGroupEntitlements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.ListGroupEntitlementsReq
		expected *v0Groups.ListGroupEntitlementsResp
		err      error
	}{
		{
			name: "Successful list",
			req:  &v0Groups.ListGroupEntitlementsReq{Id: "group1"},
			expected: &v0Groups.ListGroupEntitlementsResp{
				Data:    []string{"can_view::identity:id1"},
				Status:  http.StatusOK,
				Message: strPtr("List of entitlements"),
			},
			err: nil,
		},
		{
			name:     "Empty group ID",
			req:      &v0Groups.ListGroupEntitlementsReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name:     "Paginator load error",
			req:      &v0Groups.ListGroupEntitlementsReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to load token paginator: some error"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.ListGroupEntitlementsReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list entitlements for group group1, some error"),
		},
		{
			name:     "Pagination metadata error",
			req:      &v0Groups.ListGroupEntitlementsReq{Id: "group1"},
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

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Paginator load error" || test.name == "Pagination metadata error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
			}

			var serviceReturnValues []string
			if test.expected != nil {
				serviceReturnValues = test.expected.Data
			}

			if test.name != "Empty group ID" {
				mockService.EXPECT().ListPermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(serviceReturnValues, nil, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.ListGroupEntitlements(mockCtx, test.req)

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

func TestUpdateGroupEntitlements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.UpdateGroupEntitlementsReq
		expected *v0Groups.UpdateGroupEntitlementsResp
		err      error
	}{
		{
			name: "Successful update",
			req: &v0Groups.UpdateGroupEntitlementsReq{
				Id: "group1",
				EntitlementsPatchReq: &v0Types.Permissions{
					Updates: []*v0Types.Permission{
						{Relation: "relation1", Object: "object1"},
					},
				},
			},
			expected: &v0Groups.UpdateGroupEntitlementsResp{
				Status:  http.StatusOK,
				Message: strPtr("Updated permissions for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.UpdateGroupEntitlementsReq{
				Id: "",
				EntitlementsPatchReq: &v0Types.Permissions{
					Updates: []*v0Types.Permission{
						{Relation: "relation1", Object: "object1"},
					},
				},
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.UpdateGroupEntitlementsReq{
				Id: "group1",
				EntitlementsPatchReq: &v0Types.Permissions{
					Updates: []*v0Types.Permission{
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.UpdateGroupEntitlements").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				mockService.EXPECT().AssignPermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.UpdateGroupEntitlements(mockCtx, test.req)

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

func TestRemoveGroupEntitlement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.RemoveGroupEntitlementReq
		expected *v0Groups.RemoveGroupEntitlementResp
		err      error
	}{
		{
			name: "Successful removal",
			req: &v0Groups.RemoveGroupEntitlementReq{
				Id:            "group1",
				EntitlementId: "can_edit::client:okta",
			},
			expected: &v0Groups.RemoveGroupEntitlementResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed permission can_edit for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.RemoveGroupEntitlementReq{
				Id:            "",
				EntitlementId: "can_edit::client:okta",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Empty entitlement ID",
			req: &v0Groups.RemoveGroupEntitlementReq{
				Id:            "group1",
				EntitlementId: "",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "entitlement ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.RemoveGroupEntitlementReq{
				Id:            "group1",
				EntitlementId: "can_edit::client:okta",
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to remove permission relation1 from group group1: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.RemoveGroupEntitlement").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" || test.name == "Empty entitlement ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" && test.name != "Empty entitlement ID" {
				mockService.EXPECT().RemovePermissions(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveGroupEntitlement(mockCtx, test.req)

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

func TestGetGroupRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.GetGroupRolesReq
		expected *v0Groups.GetGroupRolesResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Groups.GetGroupRolesReq{Id: "group1"},
			expected: &v0Groups.GetGroupRolesResp{
				Data:    []string{"role1"},
				Status:  http.StatusOK,
				Message: strPtr("List of roles"),
			},
			err: nil,
		},
		{
			name:     "Empty group ID",
			req:      &v0Groups.GetGroupRolesReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.GetGroupRolesReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list group ID group1 roles: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.GetGroupRoles").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				var expectedData []string = nil
				if test.expected != nil {
					expectedData = test.expected.Data
				}
				mockService.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(expectedData, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.GetGroupRoles(mockCtx, test.req)

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

func TestUpdateGroupRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.UpdateGroupRolesReq
		expected *v0Groups.UpdateGroupRolesResp
		err      error
	}{
		{
			name: "Successful update",
			req: &v0Groups.UpdateGroupRolesReq{
				Id: "group1",
				Roles: &v0Groups.Roles{Roles: []string{
					"role1",
					"role2",
				}},
			},
			expected: &v0Groups.UpdateGroupRolesResp{
				Status:  http.StatusOK,
				Message: strPtr("Updated roles for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.UpdateGroupRolesReq{
				Id: "",
				Roles: &v0Groups.Roles{Roles: []string{
					"role1",
					"role2",
				}},
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.UpdateGroupRolesReq{
				Id: "group1",
				Roles: &v0Groups.Roles{Roles: []string{
					"role1",
					"role2",
				}},
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to assign roles: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.UpdateGroupRoles").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				mockService.EXPECT().AssignRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.UpdateGroupRoles(mockCtx, test.req)

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

func TestRemoveGroupRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.RemoveGroupRoleReq
		expected *v0Groups.RemoveGroupRoleResp
		err      error
	}{
		{
			name: "Successful removal",
			req: &v0Groups.RemoveGroupRoleReq{
				Id:     "group1",
				RoleId: "role1",
			},
			expected: &v0Groups.RemoveGroupRoleResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed role role1 for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.RemoveGroupRoleReq{
				Id:     "",
				RoleId: "role1",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Empty role ID",
			req: &v0Groups.RemoveGroupRoleReq{
				Id:     "group1",
				RoleId: "",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.RemoveGroupRoleReq{
				Id:     "group1",
				RoleId: "role1",
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to remove role role1 from group group1: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.RemoveGroupRole").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" || test.name == "Empty role ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" && test.name != "Empty role ID" {
				mockService.EXPECT().RemoveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveGroupRole(mockCtx, test.req)

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

func TestGetGroupIdentities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.GetGroupIdentitiesReq
		expected *v0Groups.GetGroupIdentitiesResp
		err      error
	}{
		{
			name: "Successful retrieval",
			req:  &v0Groups.GetGroupIdentitiesReq{Id: "group1"},
			expected: &v0Groups.GetGroupIdentitiesResp{
				Data:    []string{"identity1"},
				Status:  http.StatusOK,
				Message: strPtr("List of identities"),
			},
			err: nil,
		},
		{
			name:     "Empty group ID",
			req:      &v0Groups.GetGroupIdentitiesReq{Id: ""},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name:     "Service error",
			req:      &v0Groups.GetGroupIdentitiesReq{Id: "group1"},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to list group ID group1 identities: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.GetGroupIdentities").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				var expectedData []string = nil
				if test.expected != nil {
					expectedData = test.expected.Data
				}
				mockService.EXPECT().ListIdentities(gomock.Any(), gomock.Any()).Return(expectedData, test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.GetGroupIdentities(mockCtx, test.req)

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

func TestUpdateGroupIdentities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.UpdateGroupIdentitiesReq
		expected *v0Groups.UpdateGroupIdentitiesResp
		err      error
	}{
		{
			name: "Successful update",
			req: &v0Groups.UpdateGroupIdentitiesReq{
				Id: "group1",
				Identities: &v0Groups.Identities{
					Identities: []string{
						"identity1",
						"identity2",
					},
				},
			},
			expected: &v0Groups.UpdateGroupIdentitiesResp{
				Status:  http.StatusOK,
				Message: strPtr("Updated identities for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.UpdateGroupIdentitiesReq{
				Id: "",
				Identities: &v0Groups.Identities{
					Identities: []string{
						"identity1",
						"identity2",
					},
				},
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.UpdateGroupIdentitiesReq{
				Id: "group1",
				Identities: &v0Groups.Identities{
					Identities: []string{
						"identity1",
						"identity2",
					},
				},
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to assign identities: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.UpdateGroupIdentities").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" {
				mockService.EXPECT().AssignIdentities(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.UpdateGroupIdentities(mockCtx, test.req)

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

func TestRemoveGroupIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		req      *v0Groups.RemoveGroupIdentityReq
		expected *v0Groups.RemoveGroupIdentityResp
		err      error
	}{
		{
			name: "Successful removal",
			req: &v0Groups.RemoveGroupIdentityReq{
				Id:         "group1",
				IdentityId: "identity1",
			},
			expected: &v0Groups.RemoveGroupIdentityResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed identity identity1 for group group1"),
			},
			err: nil,
		},
		{
			name: "Empty group ID",
			req: &v0Groups.RemoveGroupIdentityReq{
				Id:         "",
				IdentityId: "identity1",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "group ID is empty"),
		},
		{
			name: "Empty identity ID",
			req: &v0Groups.RemoveGroupIdentityReq{
				Id:         "group1",
				IdentityId: "",
			},
			expected: nil,
			err:      status.Errorf(codes.InvalidArgument, "role ID is empty"),
		},
		{
			name: "Service error",
			req: &v0Groups.RemoveGroupIdentityReq{
				Id:         "group1",
				IdentityId: "identity1",
			},
			expected: nil,
			err:      status.Errorf(codes.Internal, "failed to remove identity identity1 from group group1: some error"),
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

			mockTracer.EXPECT().Start(gomock.Any(), "groups.GrpcHandler.RemoveGroupIdentity").Return(mockCtx, mockSpan)

			if test.name == "Empty group ID" || test.name == "Empty identity ID" {
				mockLogger.EXPECT().Debug(gomock.Any())
			} else if test.name == "Service error" {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			}

			if test.name != "Empty group ID" && test.name != "Empty identity ID" {
				mockService.EXPECT().RemoveIdentities(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.err)
			}

			handler := NewGrpcHandler(mockService, mockTracer, mockMonitor, mockLogger)
			resp, err := handler.RemoveGroupIdentity(mockCtx, test.req)

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
