// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"context"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	v0Types "github.com/canonical/identity-platform-api/v0/http"
	hClient "github.com/ory/hydra-client-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
)

const DEFAULT_PAGINATION_SIZE = 100

type GrpcHandler struct {
	svc    ServiceInterface
	mapper *GrpcPbMapper
	// UnimplementedClientsServiceServer must be embedded to get forward compatible implementations.
	v0Clients.UnimplementedClientsServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) ListClients(ctx context.Context, req *v0Clients.ListClientsReq) (*v0Clients.ListClientsResp, error) {
	ctx, span := g.tracer.Start(ctx, "clients.GrpcHandler.ListClients")
	defer span.End()

	var pageSize = DEFAULT_PAGINATION_SIZE
	if req.GetSize() != "" {
		size, err := strconv.Atoi(req.GetSize())
		if err != nil {
			g.logger.Error("page size parameter is not an int")
			return nil, status.Errorf(codes.InvalidArgument, "page size parameter is not an int, %v", err)
		}

		pageSize = size
	}

	r := NewListClientsRequest(req.GetClientName(), req.GetOwner(), req.GetPageToken(), pageSize)
	res, err := g.svc.ListClients(ctx, r)
	if err != nil {
		g.logger.Errorf("unexpected internal error, %v", err)
		return nil, status.Error(codes.Internal, "unexpected internal error")
	}
	if res.ServiceError != nil {
		g.logger.Errorf("failed to list clients, %v", res.ServiceError.Error)
		return nil, status.Error(g.mapper.codeFromHTTPStatus(res.ServiceError.StatusCode), "failed to list clients")
	}

	clients, ok := res.Resp.([]hClient.OAuth2Client)
	if !ok {
		g.logger.Debug("failed to convert to hydra OAuth2Client")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	data, err := g.mapper.FromOAuth2Clients(clients)
	if err != nil {
		g.logger.Errorf("failed to map from hydra OAuth2 clients, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "List of clients"

	return &v0Clients.ListClientsResp{
		Data:    data,
		Status:  http.StatusOK,
		Message: &message,
		Meta: &v0Types.Pagination{
			Size: int32(len(data)),
			Next: &res.Tokens.Next,
			Prev: &res.Tokens.Prev,
		},
	}, nil
}

func (g *GrpcHandler) GetClient(ctx context.Context, req *v0Clients.GetClientReq) (*v0Clients.GetClientResp, error) {
	ctx, span := g.tracer.Start(ctx, "clients.GrpcHandler.GetClient")
	defer span.End()

	id := req.GetId()
	res, err := g.svc.GetClient(ctx, id)
	if err != nil {
		g.logger.Errorf("unexpected internal error, %v", err)
		return nil, status.Error(codes.Internal, "unexpected internal error")
	}
	if res.ServiceError != nil {
		g.logger.Errorf("failed to get client, %v", res.ServiceError.Error)
		return nil, status.Error(g.mapper.codeFromHTTPStatus(res.ServiceError.StatusCode), "failed to get client")
	}

	client, ok := res.Resp.(*hClient.OAuth2Client)
	if !ok {
		g.logger.Debug("failed to convert to hydra OAuth2Client")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	data, err := g.mapper.fromOAuth2Client(client)
	if err != nil {
		g.logger.Errorf("failed to map from hydra OAuth2 client, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "Client detail"

	return &v0Clients.GetClientResp{
		Data:    data,
		Status:  http.StatusOK,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) CreateClient(ctx context.Context, req *v0Clients.CreateClientReq) (*v0Clients.CreateClientResp, error) {
	ctx, span := g.tracer.Start(ctx, "clients.GrpcHandler.CreateClient")
	defer span.End()

	hydraClient, err := g.mapper.ToOAuth2Client(req.GetClient())
	if err != nil {
		g.logger.Errorf("failed to map to hydra OAuth2 client, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	res, err := g.svc.CreateClient(ctx, hydraClient)
	if err != nil {
		g.logger.Errorf("unexpected internal error, %v", err)
		return nil, status.Error(codes.Internal, "unexpected internal error")
	}
	if res.ServiceError != nil {
		g.logger.Errorf("failed to create client, %v", res.ServiceError.Error)
		return nil, status.Error(g.mapper.codeFromHTTPStatus(res.ServiceError.StatusCode), "failed to create client")
	}

	client, ok := res.Resp.(*hClient.OAuth2Client)
	if !ok {
		g.logger.Debug("failed to convert to hydra OAuth2 client")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	data, err := g.mapper.fromOAuth2Client(client)
	if err != nil {
		g.logger.Errorf("failed to map from hydra OAuth2 clients, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "Created client"

	return &v0Clients.CreateClientResp{
		Data:    data,
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) UpdateClient(ctx context.Context, req *v0Clients.UpdateClientReq) (*v0Clients.UpdateClientResp, error) {
	ctx, span := g.tracer.Start(ctx, "clients.GrpcHandler.UpdateClient")
	defer span.End()

	id := req.GetId()
	if id == "" {
		g.logger.Error("client ID is empty")
		return nil, status.Error(codes.InvalidArgument, "client ID is empty")
	}

	client := req.GetClient()
	if client == nil {
		g.logger.Error("empty request body")
		return nil, status.Error(codes.InvalidArgument, "empty request body")
	}

	if client.GetClientId() == "" {
		client.ClientId = &id
	}

	if id != client.GetClientId() {
		g.logger.Errorf("client ID mismatch, %v != %v", id, client.GetClientId())
		return nil, status.Errorf(codes.InvalidArgument, "client ID mismatch, %v != %v", id, client.GetClientId())
	}

	hydraClient, err := g.mapper.ToOAuth2Client(req.Client)
	if err != nil {
		g.logger.Errorf("failed to map to hydra OAuth2 client, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	res, err := g.svc.UpdateClient(ctx, hydraClient)
	if err != nil {
		g.logger.Errorf("unexpected internal error, %v", err)
		return nil, status.Error(codes.Internal, "unexpected internal error")
	}
	if res.ServiceError != nil {
		g.logger.Errorf("failed to update client, %v", res.ServiceError.Error)
		return nil, status.Error(g.mapper.codeFromHTTPStatus(res.ServiceError.StatusCode), "failed to update client")
	}

	updatedClient, ok := res.Resp.(*hClient.OAuth2Client)
	if !ok {
		g.logger.Debug("failed to convert to hydra OAuth2 client")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	data, err := g.mapper.fromOAuth2Client(updatedClient)
	if err != nil {
		g.logger.Errorf("failed to map from hydra OAuth2 clients, %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	message := "Updated client"

	return &v0Clients.UpdateClientResp{
		Data:    data,
		Status:  http.StatusCreated,
		Message: &message,
	}, nil
}

func (g *GrpcHandler) RemoveClient(ctx context.Context, req *v0Clients.RemoveClientReq) (*v0Clients.RemoveClientResp, error) {
	ctx, span := g.tracer.Start(ctx, "clients.GrpcHandler.RemoveClient")
	defer span.End()

	id := req.GetId()
	res, err := g.svc.DeleteClient(ctx, id)
	if err != nil {
		g.logger.Errorf("unexpected internal error, %v", err)
		return nil, status.Error(codes.Internal, "unexpected internal error")
	}
	if res.ServiceError != nil {
		g.logger.Errorf("failed to remove client, %v", res.ServiceError.Error)
		return nil, status.Error(g.mapper.codeFromHTTPStatus(res.ServiceError.StatusCode), "failed to remove client")
	}

	message := "Removed client"

	return &v0Clients.RemoveClientResp{
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
