// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
	"encoding/json"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
)

type GrpcHandler struct {
	svc    ServiceInterface
	mapper *GrpcPbMapper
	// UnimplementedIdpsServiceServer must be embedded to get forward compatible implementations.
	v0Idps.UnimplementedIdpsServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListIdps(ctx context.Context, _ *emptypb.Empty) (*v0Idps.ListIdpsResp, error) {
	ctx, span := g.tracer.Start(ctx, "idps.GrpcHandler.ListIdps")
	defer span.End()

	configurations, err := g.svc.ListResources(ctx)
	if err != nil {
		g.logger.Errorf("failed to list IDPs: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list IDPs: %v", err)
	}

	idps, err := g.mapper.FromConfigurations(configurations)
	if err != nil {
		g.logger.Errorf("failed to map from kratos identity provider configs, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "List of IDPs"
	return &v0Idps.ListIdpsResp{
		Data:    idps,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetIdp(ctx context.Context, req *v0Idps.GetIdpReq) (*v0Idps.GetIdpResp, error) {
	ctx, span := g.tracer.Start(ctx, "idps.GrpcHandler.GetIdp")
	defer span.End()

	id := req.GetId()

	configurations, err := g.svc.GetResource(ctx, req.GetId())
	if err != nil {
		g.logger.Errorf("failed to get IDP %s, %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get IDP %s, %v", id, err)
	}

	idps, err := g.mapper.FromConfigurations(configurations)
	if err != nil {
		g.logger.Errorf("failed to map from kratos identity provider configs, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "IDP details"
	return &v0Idps.GetIdpResp{
		Data:    idps,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateIdp(ctx context.Context, req *v0Idps.CreateIdpReq) (*v0Idps.CreateIdpResp, error) {
	ctx, span := g.tracer.Start(ctx, "idps.GrpcHandler.CreateIdp")
	defer span.End()

	createIdpBody := req.GetIdp()
	if createIdpBody == nil {
		g.logger.Error("empty request body")
		return nil, status.Errorf(codes.InvalidArgument, "empty request body")
	}
	requestedClaims := createIdpBody.GetRequestedClaims()
	if requestedClaims != "" && !json.Valid([]byte(requestedClaims)) {
		g.logger.Errorf("invalid requested claims: %v", requestedClaims)
		return nil, status.Errorf(codes.InvalidArgument, "invalid requested claims: %v", requestedClaims)
	}

	configuration, err := g.mapper.ToCreateIdpBody(createIdpBody)
	if err != nil {
		g.logger.Errorf("failed to map to kratos identity provider config, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map to kratos identity provider config, %v", err)
	}

	configurations, err := g.svc.CreateResource(ctx, configuration)
	if err != nil {
		g.logger.Errorf("failed to create IDP, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create IDP, %v", err)
	}

	idps, err := g.mapper.FromConfigurations(configurations)
	if err != nil {
		g.logger.Errorf("failed to map from kratos identity provider configs, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "Created IDP"
	return &v0Idps.CreateIdpResp{
		Data:    idps,
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateIdp(ctx context.Context, req *v0Idps.UpdateIdpReq) (*v0Idps.UpdateIdpResp, error) {
	ctx, span := g.tracer.Start(ctx, "idps.GrpcHandler.UpdateIdp")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Error("IDP ID is empty")
		return nil, status.Error(codes.InvalidArgument, "IDP ID is empty")
	}

	updateIdpBody := req.GetIdp()
	if updateIdpBody == nil {
		g.logger.Error("empty request body")
		return nil, status.Error(codes.InvalidArgument, "empty request body")
	}
	if id != updateIdpBody.GetId() {
		g.logger.Errorf("IDP ID mismatch, %v != %v", id, updateIdpBody.GetId())
		return nil, status.Errorf(codes.InvalidArgument, "IDP ID mismatch, %v != %v", id, updateIdpBody.GetId())
	}

	requestedClaims := updateIdpBody.GetRequestedClaims()
	if requestedClaims != "" && !json.Valid([]byte(requestedClaims)) {
		g.logger.Errorf("invalid requested claims: %v", requestedClaims)
		return nil, status.Errorf(codes.InvalidArgument, "invalid requested claims: %v", requestedClaims)
	}

	configuration, err := g.mapper.ToUpdateIdpBody(updateIdpBody)
	if err != nil {
		g.logger.Errorf("failed to map to createIdpBody, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	configurations, err := g.svc.EditResource(ctx, id, configuration)
	if configurations != nil && len(configurations) == 0 {
		return nil, status.Errorf(codes.NotFound, "IDP %v not found", id)
	}
	if err != nil {
		g.logger.Errorf("failed to update IDP, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update IDP, %v", err)
	}

	idps, err := g.mapper.FromConfigurations(configurations)
	if err != nil {
		g.logger.Errorf("failed to map from kratos identity provider configs, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "Updated IDP"
	return &v0Idps.UpdateIdpResp{
		Data:    idps,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveIdp(ctx context.Context, req *v0Idps.RemoveIdpReq) (*v0Idps.RemoveIdpResp, error) {
	ctx, span := g.tracer.Start(ctx, "idps.GrpcHandler.RemoveIdp")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Debug("IDP ID is empty")
		return nil, status.Error(codes.InvalidArgument, "IDP ID is empty")
	}

	err := g.svc.DeleteResource(ctx, id)
	if err != nil {
		g.logger.Errorf("failed to remove IDP, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to remove IDP, %v", err)
	}

	message := "Removed IDP"
	return &v0Idps.RemoveIdpResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func NewGrpcHandler(svc ServiceInterface, mapper *GrpcPbMapper, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GrpcHandler {
	return &GrpcHandler{
		svc:     svc,
		mapper:  mapper,
		logger:  logger,
		tracer:  tracer,
		monitor: monitor,
	}
}
