// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"context"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	v0Types "github.com/canonical/identity-platform-api/v0/http"
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"strconv"
)

const DEFAULT_PAGINATION_SIZE = 100

type GrpcHandler struct {
	svc    ServiceInterface
	mapper *GrpcPbMapper
	// UnimplementedSchemasServiceServer must be embedded to get forward compatible implementations.
	v0Schemas.UnimplementedSchemasServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListSchemas(ctx context.Context, req *v0Schemas.ListSchemasReq) (*v0Schemas.ListSchemasResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.ListSchemas")
	defer span.End()

	var pageSize int64 = DEFAULT_PAGINATION_SIZE
	if req.GetSize() != "" {
		size, err := strconv.ParseInt(req.GetSize(), 10, 64)
		if err != nil {
			g.logger.Error("page size parameter is not an int")
			return nil, status.Errorf(codes.InvalidArgument, "page size parameter is not an int, %v", err)
		}

		pageSize = size
	}

	schemas, err := g.svc.ListSchemas(ctx, pageSize, req.GetPageToken())
	if err != nil {
		g.logger.Errorf("failed to list schemas, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list schemas, %v", err)
	}

	mappedSchemas, err := g.mapper.FromIdentitySchemaContainerModel(schemas.IdentitySchemas)
	if err != nil {
		g.logger.Errorf("failed to map from kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from kratos schema containers, %v", err)
	}

	message := "List of schemas"

	next := schemas.Tokens.Next
	prev := schemas.Tokens.Prev

	return &v0Schemas.ListSchemasResp{
		Data:    mappedSchemas,
		Status:  http.StatusOK,
		Message: &message,
		Meta: &v0Types.Pagination{
			Size: int32(len(schemas.IdentitySchemas)),
			Next: &next,
			Prev: &prev,
		},
	}, nil
}

func (g *GrpcHandler) GetSchema(ctx context.Context, req *v0Schemas.GetSchemaReq) (*v0Schemas.GetSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.GetSchema")
	defer span.End()

	id := req.GetId()

	schemaData, err := g.svc.GetSchema(ctx, id)
	if schemaData != nil && len(schemaData.IdentitySchemas) == 0 {
		return nil, status.Errorf(codes.NotFound, "schema %s not found", id)
	}
	if err != nil {
		g.logger.Errorf("failed to get schema %s, %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get schema %s, %v", id, err)
	}

	mappedSchemas, err := g.mapper.FromIdentitySchemaContainerModel(schemaData.IdentitySchemas)
	if err != nil {
		g.logger.Errorf("failed to map from kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from kratos schema containers, %v", err)
	}

	message := "Schema detail"

	return &v0Schemas.GetSchemaResp{
		Data:    mappedSchemas,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateSchema(ctx context.Context, req *v0Schemas.CreateSchemaReq) (*v0Schemas.CreateSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.CreateSchema")
	defer span.End()

	body := req.GetSchema()
	schemaContainer, err := g.mapper.ToIdentitySchemaContainerModel(body)
	if err != nil {
		g.logger.Errorf("failed to map to kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map to kratos schema containers, %v", err)
	}

	schemaData, err := g.svc.CreateSchema(ctx, schemaContainer)
	if err != nil {
		g.logger.Errorf("failed to create schema, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create schema, %v", err)
	}

	mappedSchemas, err := g.mapper.FromIdentitySchemaContainerModel(schemaData.IdentitySchemas)
	if err != nil {
		g.logger.Errorf("failed to map from kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from kratos schema containers, %v", err)
	}

	message := "Created schema"

	return &v0Schemas.CreateSchemaResp{
		Data:    mappedSchemas,
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateSchema(ctx context.Context, req *v0Schemas.UpdateSchemaReq) (*v0Schemas.UpdateSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.UpdateSchema")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Error("schema ID is empty")
		return nil, status.Error(codes.InvalidArgument, "schema ID is empty")
	}

	body := req.GetSchema()
	if body == nil {
		g.logger.Error("empty request body")
		return nil, status.Error(codes.InvalidArgument, "empty request body")
	}
	if id != body.GetId() {
		g.logger.Errorf("schema ID mismatch, %v != %v", id, body.GetId())
		return nil, status.Errorf(codes.InvalidArgument, "schema ID mismatch, %v != %v", id, body.GetId())
	}

	schemaContainer, err := g.mapper.ToIdentitySchemaContainerModel(body)
	if err != nil {
		g.logger.Errorf("failed to map to kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map to kratos schema containers, %v", err)
	}

	schemaData, err := g.svc.EditSchema(ctx, id, schemaContainer)
	if schemaData != nil && len(schemaData.IdentitySchemas) == 0 {
		return nil, status.Errorf(codes.NotFound, "schema %s not found", id)
	}
	if err != nil {
		g.logger.Errorf("failed to update schema, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update schema, %v", err)
	}

	mappedSchema, err := g.mapper.FromIdentitySchemaContainerModel(schemaData.IdentitySchemas)
	if err != nil {
		g.logger.Errorf("failed to map from kratos schema containers, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to map from kratos schema containers, %v", err)
	}

	message := "Updated schema"
	return &v0Schemas.UpdateSchemaResp{
		Data:    mappedSchema,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveSchema(ctx context.Context, req *v0Schemas.RemoveSchemaReq) (*v0Schemas.RemoveSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.RemoveSchema")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Error("schema ID is empty")
		return nil, status.Error(codes.InvalidArgument, "schema ID is empty")
	}

	err := g.svc.DeleteSchema(ctx, id)
	if err != nil {
		g.logger.Errorf("failed to remove schema, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to remove schema, %v", err)
	}

	message := "Removed schema"
	return &v0Schemas.RemoveSchemaResp{
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) GetDefaultSchema(ctx context.Context, _ *emptypb.Empty) (*v0Schemas.GetDefaultSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.GetDefaultSchema")
	defer span.End()

	defaultSchema, err := g.svc.GetDefaultSchema(ctx)
	if err != nil {
		g.logger.Errorf("failed to get default schema, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get default schema, %v", err)
	}

	message := "Get default schema"
	return &v0Schemas.GetDefaultSchemaResp{
		Data:    &v0Schemas.DefaultSchema{Id: defaultSchema.ID},
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateDefaultSchema(ctx context.Context, req *v0Schemas.UpdateDefaultSchemaReq) (*v0Schemas.UpdateDefaultSchemaResp, error) {
	ctx, span := g.tracer.Start(ctx, "schemas.GrpcHandler.UpdateDefaultSchema")
	defer span.End()

	if req.GetSchema() == nil {
		return nil, status.Error(codes.InvalidArgument, "schema is empty")
	}

	defaultSchema, err := g.svc.UpdateDefaultSchema(ctx, &DefaultSchema{ID: req.GetSchema().GetId()})
	if err != nil {
		g.logger.Errorf("failed to update default schema, %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update default schema, %v", err)
	}

	message := "Updated default schema"
	return &v0Schemas.UpdateDefaultSchemaResp{
		Data:    &v0Schemas.DefaultSchema{Id: defaultSchema.ID},
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func NewGrpcHandler(svc ServiceInterface, mapper *GrpcPbMapper, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GrpcHandler {
	return &GrpcHandler{
		svc:     svc,
		mapper:  mapper,
		tracer:  tracer,
		monitor: monitor,
		logger:  logger,
	}
}
