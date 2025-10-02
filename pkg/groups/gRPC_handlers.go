// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

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

	v0Groups "github.com/canonical/identity-platform-api/v0/groups"
)

type GrpcHandler struct {
	svc ServiceInterface
	// UnimplementedGroupsServiceServer must be embedded to get forward compatible implementations.
	v0Groups.UnimplementedGroupsServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListGroups(ctx context.Context, _ *emptypb.Empty) (*v0Groups.ListGroupsResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.ListGroups")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)

	roles, err := g.svc.ListGroups(
		ctx,
		principal.Identifier(),
	)

	if err != nil {
		g.logger.Errorf("failed to list groups: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list groups: %v", err)
	}

	message := "List of groups"

	return &v0Groups.ListGroupsResp{
		Data:    roles,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetGroup(ctx context.Context, req *v0Groups.GetGroupReq) (*v0Groups.GetGroupResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.GetGroup")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)
	group, err := g.svc.GetGroup(ctx, principal.Identifier(), req.GetId())

	if err != nil {
		g.logger.Errorf("failed to retrieve group: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve group: %v", err)
	}

	if group == nil {
		g.logger.Debugf("group not found: %v", req.GetId())
		return nil, status.Errorf(codes.NotFound, "group not found: %v", req.GetId())
	}

	message := "Group detail"

	return &v0Groups.GetGroupResp{
		Data: []*v0Groups.Group{
			{
				Id:   group.ID,
				Name: group.Name,
			},
		},
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateGroup(ctx context.Context, req *v0Groups.CreateGroupReq) (*v0Groups.CreateGroupResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.CreateGroup")
	defer span.End()

	role := req.GetGroup()
	if role == nil || role.GetName() == "" {
		g.logger.Debug("group name is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group name is empty")
	}

	principal := authentication.PrincipalFromContext(ctx)
	createdRole, err := g.svc.CreateGroup(ctx, principal.Identifier(), role.GetName())

	if err != nil {
		g.logger.Errorf("failed to create group: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}

	message := fmt.Sprintf("Created group %s", createdRole.Name)

	return &v0Groups.CreateGroupResp{
		Data: []*v0Groups.Group{
			{
				Id:   createdRole.ID,
				Name: createdRole.Name,
			},
		},
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateGroup(ctx context.Context, _ *v0Groups.UpdateGroupReq) (*v0Groups.UpdateGroupResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.UpdateGroup")
	defer span.End()

	return nil, status.Error(codes.Unimplemented, "method not implemented")
}

func (g *GrpcHandler) RemoveGroup(ctx context.Context, req *v0Groups.RemoveGroupReq) (*v0Groups.RemoveGroupResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.RemoveGroup")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	if err := g.svc.DeleteGroup(ctx, groupID); err != nil {
		g.logger.Errorf("failed to delete role: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	message := fmt.Sprintf("Deleted group %s", groupID)

	return &v0Groups.RemoveGroupResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) ListGroupEntitlements(ctx context.Context, req *v0Groups.ListGroupEntitlementsReq) (*v0Groups.ListGroupEntitlementsResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.ListGroupEntitlements")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	paginator := types.NewTokenPaginator(g.tracer, g.logger)
	if err := paginator.LoadFromGRPCContext(ctx); err != nil {
		g.logger.Errorf("failed to load token paginator: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to load token paginator: %v", err)
	}

	permissions, pageTokens, err := g.svc.ListPermissions(ctx, groupID, paginator.GetAllTokens(ctx))
	if err != nil {
		g.logger.Errorf("failed to list entitlements for group %s, %v", groupID, err)
		return nil, status.Errorf(codes.Internal, "failed to list entitlements for group %s, %v", groupID, err)
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

	return &v0Groups.ListGroupEntitlementsResp{
		Data:    permissions,
		Status:  http.StatusOK,
		Message: &message,
	}, nil

}

func (g *GrpcHandler) UpdateGroupEntitlements(ctx context.Context, req *v0Groups.UpdateGroupEntitlementsReq) (*v0Groups.UpdateGroupEntitlementsResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.UpdateGroupEntitlements")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	permissions := make([]Permission, 0, len(req.GetEntitlementsPatchReq().GetUpdates()))

	for _, permission := range req.GetEntitlementsPatchReq().GetUpdates() {
		permissions = append(permissions, Permission{
			Relation: permission.GetRelation(),
			Object:   permission.GetObject(),
		})
	}

	if err := g.svc.AssignPermissions(ctx, groupID, permissions...); err != nil {
		g.logger.Errorf("failed to assign permissions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to assign permissions: %v", err)
	}

	message := fmt.Sprintf("Updated permissions for group %s", groupID)

	return &v0Groups.UpdateGroupEntitlementsResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveGroupEntitlement(ctx context.Context, req *v0Groups.RemoveGroupEntitlementReq) (*v0Groups.RemoveGroupEntitlementResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.RemoveGroupEntitlement")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	entitlementID := req.GetEntitlementId()
	if entitlementID == "" {
		g.logger.Debug("entitlement ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "entitlement ID is empty")
	}

	permission := authorization.NewURNFromURLParam(entitlementID)

	err := g.svc.RemovePermissions(ctx, groupID, Permission{Relation: permission.Relation(), Object: permission.Object()})
	if err != nil {
		g.logger.Errorf("failed to remove permission %s from group %s: %v", permission.Relation(), groupID, err)
		return nil, status.Errorf(
			codes.Internal,
			"failed to remove permission %s from group %s: %v",
			permission.Relation(), groupID, err,
		)
	}

	message := fmt.Sprintf("Removed permission %s for group %s", permission.Relation(), groupID)

	return &v0Groups.RemoveGroupEntitlementResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetGroupRoles(ctx context.Context, req *v0Groups.GetGroupRolesReq) (*v0Groups.GetGroupRolesResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.GetGroupRoles")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	groups, err := g.svc.ListRoles(ctx, groupID)
	if err != nil {
		g.logger.Errorf("failed to list group ID %s roles: %v", groupID, err)
		return nil, status.Errorf(codes.Internal, "failed to list group ID %s roles: %v", groupID, err)
	}

	message := "List of roles"

	return &v0Groups.GetGroupRolesResp{
		Data:    groups,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateGroupRoles(ctx context.Context, req *v0Groups.UpdateGroupRolesReq) (*v0Groups.UpdateGroupRolesResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.UpdateGroupRoles")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	roles := req.GetRoles().GetRoles()
	principal := authentication.PrincipalFromContext(ctx)
	canAssign, err := g.svc.CanAssignRoles(ctx, principal.Identifier(), roles...)
	if err != nil {
		g.logger.Errorf("failed to check if roles can be assigned: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check if roles can be assigned: %v", err)
	}
	if !canAssign {
		return nil, status.Errorf(
			codes.PermissionDenied,
			"user %s is not allowed to assign specified roles",
			principal.Identifier(),
		)
	}

	if err := g.svc.AssignRoles(ctx, groupID, roles...); err != nil {
		g.logger.Errorf("failed to assign roles: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to assign roles: %v", err)
	}

	message := fmt.Sprintf("Updated roles for group %s", groupID)

	return &v0Groups.UpdateGroupRolesResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveGroupRole(ctx context.Context, req *v0Groups.RemoveGroupRoleReq) (*v0Groups.RemoveGroupRoleResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.RemoveGroupRole")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	roleID := req.GetRoleId()
	if roleID == "" {
		g.logger.Debug("role ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "role ID is empty")
	}

	err := g.svc.RemoveRoles(ctx, groupID, roleID)
	if err != nil {
		g.logger.Errorf("failed to remove role %s for group ID %s: %v", roleID, groupID, err)
		return nil, status.Errorf(codes.Internal, "failed to remove role %s for group ID %s: %v", roleID, groupID, err)
	}

	message := fmt.Sprintf("Removed role %s for group %s", roleID, groupID)

	return &v0Groups.RemoveGroupRoleResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetGroupIdentities(ctx context.Context, req *v0Groups.GetGroupIdentitiesReq) (*v0Groups.GetGroupIdentitiesResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.GetGroupIdentities")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	identities, err := g.svc.ListIdentities(ctx, groupID)
	if err != nil {
		g.logger.Errorf("failed to list group ID %s identities: %v", groupID, err)
		return nil, status.Errorf(codes.Internal, "failed to list group ID %s identities: %v", groupID, err)
	}

	message := "List of identities"

	return &v0Groups.GetGroupIdentitiesResp{
		Data:    identities,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateGroupIdentities(ctx context.Context, req *v0Groups.UpdateGroupIdentitiesReq) (*v0Groups.UpdateGroupIdentitiesResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.UpdateGroupIdentities")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	identities := req.GetIdentities().GetIdentities()

	principal := authentication.PrincipalFromContext(ctx)
	canAssign, err := g.svc.CanAssignIdentities(ctx, principal.Identifier(), identities...)
	if err != nil {
		g.logger.Errorf("failed to check if identities can be assigned: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check if identities can be assigned: %v", err)
	}
	if !canAssign {
		return nil, status.Errorf(
			codes.PermissionDenied,
			"user %s is not allowed to assign specified identities",
			principal.Identifier(),
		)
	}

	if err := g.svc.AssignIdentities(ctx, groupID, identities...); err != nil {
		g.logger.Errorf("failed to assign identities: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to assign identities: %v", err)
	}

	message := fmt.Sprintf("Updated identities for group %s", groupID)

	return &v0Groups.UpdateGroupIdentitiesResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveGroupIdentity(ctx context.Context, req *v0Groups.RemoveGroupIdentityReq) (*v0Groups.RemoveGroupIdentityResp, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GrpcHandler.RemoveGroupIdentity")
	defer span.End()

	groupID := req.GetId()
	if groupID == "" {
		g.logger.Debug("group ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "group ID is empty")
	}

	identityID := req.GetIdentityId()
	if identityID == "" {
		g.logger.Debug("identity ID is empty")
		return nil, status.Errorf(codes.InvalidArgument, "identity ID is empty")
	}

	err := g.svc.RemoveIdentities(ctx, groupID, identityID)
	if err != nil {
		g.logger.Errorf("failed to remove identity %s for group ID %s: %v", identityID, groupID, err)
		return nil, status.Errorf(codes.Internal, "failed to remove identity %s for group ID %s: %v", identityID, groupID, err)
	}

	message := fmt.Sprintf("Removed identity %s for group %s", identityID, groupID)

	return &v0Groups.RemoveGroupIdentityResp{
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
