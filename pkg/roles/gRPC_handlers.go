// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"

	v0Roles "github.com/canonical/identity-platform-api/v0/roles"
)

type GrpcHandler struct {
	svc ServiceInterface
	// UnimplementedRolesServiceServer must be embedded to get forward compatible implementations.
	v0Roles.UnimplementedRolesServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListRoles(ctx context.Context, _ *emptypb.Empty) (*v0Roles.ListRolesResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.ListRoles")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)

	roles, err := g.svc.ListRoles(
		ctx,
		principal.Identifier(),
	)

	if err != nil {
		g.logger.Errorf("failed to list roles: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	message := "List of roles"

	return &v0Roles.ListRolesResp{
		Data:    roles,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetRole(ctx context.Context, req *v0Roles.GetRoleReq) (*v0Roles.GetRoleResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.GetRole")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)
	role, err := g.svc.GetRole(ctx, principal.Identifier(), req.GetId())

	if err != nil {
		g.logger.Errorf("failed to retrieve role: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve role: %v", err)
	}

	if role == nil {
		g.logger.Debugf("role not found: %v", req.GetId())
		return nil, status.Errorf(codes.NotFound, "role not found: %v", req.GetId())
	}

	message := "Role detail"

	return &v0Roles.GetRoleResp{
		Data: []*v0Roles.Role{
			{
				Id:   role.ID,
				Name: role.Name,
			},
		},
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateRole(ctx context.Context, req *v0Roles.CreateRoleReq) (*v0Roles.CreateRoleResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.CreateRole")
	defer span.End()

	role := req.GetRole()
	if role == nil || role.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "role name is empty")
	}

	principal := authentication.PrincipalFromContext(ctx)
	createdRole, err := g.svc.CreateRole(ctx, principal.Identifier(), role.GetName())

	if err != nil {
		g.logger.Errorf("failed to create role: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	message := fmt.Sprintf("Created role %s", createdRole.Name)

	return &v0Roles.CreateRoleResp{
		Data: []*v0Roles.Role{
			{
				Id:   createdRole.ID,
				Name: createdRole.Name,
			},
		},
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateRole(ctx context.Context, _ *v0Roles.UpdateRoleReq) (*v0Roles.UpdateRoleResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.UpdateRole")
	defer span.End()

	return nil, status.Errorf(codes.Unimplemented, "method UpdateRole not implemented")
}

func (g *GrpcHandler) RemoveRole(ctx context.Context, req *v0Roles.RemoveRoleReq) (*v0Roles.RemoveRoleResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.RemoveRole")
	defer span.End()

	roleID := req.GetId()
	if roleID == "" {
		g.logger.Debugf("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	if err := g.svc.DeleteRole(ctx, roleID); err != nil {
		g.logger.Errorf("failed to delete role: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	message := fmt.Sprintf("Deleted role %s", roleID)

	return &v0Roles.RemoveRoleResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) ListRoleEntitlements(ctx context.Context, req *v0Roles.ListRoleEntitlementsReq) (*v0Roles.ListRoleEntitlementsResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.ListRoleEntitlements")
	defer span.End()

	roleID := req.GetId()
	if roleID == "" {
		g.logger.Debugf("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	paginator := types.NewTokenPaginator(g.tracer, g.logger)

	if err := paginator.LoadFromGRPCContext(ctx); err != nil {
		g.logger.Errorf("failed to load token paginator: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to load token paginator: %v", err)
	}

	permissions, pageTokens, err := g.svc.ListPermissions(ctx, roleID, paginator.GetAllTokens(ctx))

	if err != nil {
		g.logger.Errorf("failed to list entitlements for role %s, %v", roleID, err)
		return nil, status.Errorf(codes.Internal, "failed to list entitlements for role %s, %v", roleID, err)
	}

	paginator.SetTokens(ctx, pageTokens)
	paginationMetadataValue, err := paginator.PaginationHeader(ctx)

	if err != nil {
		g.logger.Errorf("error producing pagination metadata: %v", err)
		return nil, status.Errorf(codes.Internal, "error producing pagination metadata: %v", err)
	}

	if paginationMetadataValue != "" {
		md := metadata.New(map[string]string{types.GRPC_PAGINATION_METADATA: paginationMetadataValue})
		if err = grpc.SendHeader(ctx, md); err != nil {
			g.logger.Errorf("error getting outgoing context metadata: %v", err)
			return nil, status.Errorf(codes.Internal, "error getting outgoing context metadata: %v", err)
		}
	}

	message := "List of entitlements"

	return &v0Roles.ListRoleEntitlementsResp{
		Data:    permissions,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateRoleEntitlements(ctx context.Context, req *v0Roles.UpdateRoleEntitlementsReq) (*v0Roles.UpdateRoleEntitlementsResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.UpdateRoleEntitlements")
	defer span.End()

	roleID := req.GetId()
	if roleID == "" {
		g.logger.Debugf("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	permissions := make([]Permission, 0, len(req.GetEntitlementsPatchReq().GetUpdates()))

	for _, permission := range req.GetEntitlementsPatchReq().GetUpdates() {
		permissions = append(permissions, Permission{
			Relation: permission.GetRelation(),
			Object:   permission.GetObject(),
		})
	}

	if err := g.svc.AssignPermissions(ctx, roleID, permissions...); err != nil {
		g.logger.Errorf("failed to assign permissions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to assign permissions: %v", err)
	}

	message := fmt.Sprintf("Updated permissions for role %s", roleID)

	return &v0Roles.UpdateRoleEntitlementsResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveRoleEntitlement(ctx context.Context, req *v0Roles.RemoveRoleEntitlementReq) (*v0Roles.RemoveRoleEntitlementResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.RemoveRoleEntitlement")
	defer span.End()

	roleID := req.GetId()
	if roleID == "" {
		g.logger.Errorf("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	entitlementID := req.GetEntitlementId()
	if entitlementID == "" {
		g.logger.Errorf("entitlement ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "entitlement ID is empty")
	}

	permission := authorization.NewURNFromURLParam(entitlementID)

	err := g.svc.RemovePermissions(ctx, roleID, Permission{Relation: permission.Relation(), Object: permission.Object()})
	if err != nil {
		g.logger.Errorf("failed to remove permission %s from role %s: %v", permission.Relation(), roleID, err)
		return nil, status.Errorf(
			codes.Internal,
			"failed to remove permission %s from role %s: %v",
			permission.Relation(), roleID, err,
		)
	}

	message := fmt.Sprintf("Removed permission %s for role %s", permission.Relation(), roleID)

	return &v0Roles.RemoveRoleEntitlementResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetRoleGroups(ctx context.Context, req *v0Roles.GetRoleGroupsReq) (*v0Roles.GetRoleGroupsResp, error) {
	ctx, span := g.tracer.Start(ctx, "roles.GrpcHandler.GetRoleGroups")
	defer span.End()

	roleID := req.GetId()
	if roleID == "" {
		g.logger.Debugf("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	roles, err := g.svc.ListRoleGroups(ctx, roleID)
	if err != nil {
		g.logger.Errorf("failed to list role ID %s groups: %v", roleID, err)
		return nil, status.Errorf(codes.Internal, "failed to list role ID %s groups: %v", roleID, err)
	}

	message := "List of groups"

	return &v0Roles.GetRoleGroupsResp{
		Data:    roles,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func NewGrpcHandler(svc ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GrpcHandler {
	g := new(GrpcHandler)

	g.svc = svc
	g.tracer = tracer
	g.monitor = monitor
	g.logger = logger

	return g
}
