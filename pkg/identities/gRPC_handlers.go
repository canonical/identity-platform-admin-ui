// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"net/http"
	"strconv"

	v0Types "github.com/canonical/identity-platform-api/v0/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"

	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
)

const DEFAULT_PAGINATION_SIZE = 100

type GrpcHandler struct {
	svc    ServiceInterface
	mapper *GrpcPbMapper
	// UnimplementedIdentitiesServiceServer must be embedded to get forward compatible implementations.
	v0Identities.UnimplementedIdentitiesServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListIdentities(ctx context.Context, req *v0Identities.ListIdentitiesReq) (*v0Identities.ListIdentitiesResp, error) {
	ctx, span := g.tracer.Start(ctx, "identities.GrpcHandler.ListIdentities")
	defer span.End()

	credID := req.GetCredID()
	if credID == "" {
		g.logger.Debug("credID is empty")
		return nil, status.Error(codes.InvalidArgument, "credID is empty")
	}

	var pageSize int64 = DEFAULT_PAGINATION_SIZE
	if req.GetSize() != "" {
		size, err := strconv.ParseInt(req.GetSize(), 10, 64)
		if err != nil {
			g.logger.Debug("page size parameter is not an int")
			return nil, status.Errorf(codes.InvalidArgument, "page size parameter is not an int, %v", err)
		}

		pageSize = size
	}

	identities, err := g.svc.ListIdentities(ctx, pageSize, req.GetPageToken(), credID)
	if err != nil {
		g.logger.Errorf("failed to list identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list identities, %v", err)
	}

	mappedIdentities, err := g.mapper.FromIdentitiesModel(identities.Identities)
	if err != nil {
		g.logger.Errorf("failed to map from identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from identities, %v", err)
	}

	message := "List of identities"

	next := identities.Tokens.Next
	prev := identities.Tokens.Prev

	return &v0Identities.ListIdentitiesResp{
		Data:    mappedIdentities,
		Status:  http.StatusOK,
		Message: &message,
		Meta: &v0Types.Pagination{
			Size: int32(len(identities.Identities)),
			Next: &next,
			Prev: &prev,
		},
	}, nil
}

func (g *GrpcHandler) GetIdentity(ctx context.Context, req *v0Identities.GetIdentityReq) (*v0Identities.GetIdentityResp, error) {
	ctx, span := g.tracer.Start(ctx, "identities.GrpcHandler.GetIdentity")
	defer span.End()

	id := req.GetId()

	identities, err := g.svc.GetIdentity(ctx, id)
	if err != nil {
		g.logger.Errorf("failed to get identity %s, %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get identity %s, %v", id, err)
	}

	mappedIdentities, err := g.mapper.FromIdentitiesModel(identities.Identities)
	if err != nil {
		g.logger.Errorf("failed to map from identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from identities, %v", err)
	}

	message := "Identity detail"

	return &v0Identities.GetIdentityResp{
		Data:    mappedIdentities,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateIdentity(ctx context.Context, req *v0Identities.CreateIdentityReq) (*v0Identities.CreateIdentityResp, error) {
	ctx, span := g.tracer.Start(ctx, "identities.GrpcHandler.CreateIdentity")
	defer span.End()

	createIdentityBody := req.GetIdentity()

	mappedCreateIdentityBody, err := g.mapper.ToCreateIdentityModel(createIdentityBody)
	if err != nil {
		g.logger.Errorf("failed to map to createIdentityBody, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map to createIdentityBody, %v", err)
	}

	identities, err := g.svc.CreateIdentity(ctx, mappedCreateIdentityBody)
	if err != nil {
		g.logger.Errorf("failed to create identity, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create identity, %v", err)
	}

	err = g.svc.SendUserCreationEmail(ctx, &identities.Identities[0])
	if err != nil {
		g.logger.Errorf("failed to send user creation email, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to send user creation email, %v", err)
	}

	mappedIdentities, err := g.mapper.FromIdentitiesModel(identities.Identities)
	if err != nil {
		g.logger.Errorf("failed to map from identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from identities, %v", err)
	}

	message := "Created identity"

	return &v0Identities.CreateIdentityResp{
		Data:    mappedIdentities,
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateIdentity(ctx context.Context, req *v0Identities.UpdateIdentityReq) (*v0Identities.UpdateIdentityResp, error) {
	ctx, span := g.tracer.Start(ctx, "identities.GrpcHandler.UpdateIdentity")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Debug("identity ID is empty")
		return nil, status.Error(codes.InvalidArgument, "identity ID is empty")
	}

	updateIdentityBody := req.GetIdentity()

	mappedUpdateIdentityBody, err := g.mapper.ToUpdateIdentityModel(updateIdentityBody)
	if err != nil {
		g.logger.Errorf("failed to map to updateIdentityBody, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map to updateIdentityBody, %v", err)
	}

	identities, err := g.svc.UpdateIdentity(ctx, id, mappedUpdateIdentityBody)
	if err != nil {
		g.logger.Errorf("failed to update identity %s, %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to update identity %s, %v", id, err)
	}

	mappedIdentities, err := g.mapper.FromIdentitiesModel(identities.Identities)
	if err != nil {
		g.logger.Errorf("failed to map from identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from identities, %v", err)
	}

	message := "Updated identity"

	return &v0Identities.UpdateIdentityResp{
		Data:    mappedIdentities,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveIdentity(ctx context.Context, req *v0Identities.RemoveIdentityReq) (*v0Identities.RemoveIdentityResp, error) {
	ctx, span := g.tracer.Start(ctx, "identities.GrpcHandler.RemoveIdentity")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Debug("identity ID is empty")
		return nil, status.Error(codes.InvalidArgument, "identity ID is empty")
	}

	identities, err := g.svc.DeleteIdentity(ctx, id)
	if err != nil {
		g.logger.Errorf("failed to delete identity %s, %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to delete identity %s, %v", id, err)
	}

	mappedIdentities, err := g.mapper.FromIdentitiesModel(identities.Identities)
	if err != nil {
		g.logger.Errorf("failed to map from identities, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from identities, %v", err)
	}

	message := "Identity deleted"

	return &v0Identities.RemoveIdentityResp{
		Data:    mappedIdentities,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func NewGrpcHandler(svc ServiceInterface, mapper *GrpcPbMapper, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GrpcHandler {
	g := new(GrpcHandler)

	g.svc = svc
	g.mapper = mapper

	g.tracer = tracer
	g.monitor = monitor
	g.logger = logger

	return g
}
