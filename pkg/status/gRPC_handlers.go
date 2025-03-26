// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package status

import (
	"context"

	v0Status "github.com/canonical/identity-platform-api/v0/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type GrpcHandler struct {
	// UnimplementedStatusServiceServer must be embedded to get forward compatible implementations.
	v0Status.UnimplementedStatusServiceServer

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GrpcHandler) GetStatus(ctx context.Context, _ *emptypb.Empty) (*v0Status.Status, error) {
	ctx, span := g.tracer.Start(ctx, "status.GrpcHandler.GetStatus")
	defer span.End()

	status := &v0Status.Status{
		Status:    "Ok",
		BuildInfo: v0StatusBuildInfo(),
	}

	return status, nil
}

func (g *GrpcHandler) GetVersion(ctx context.Context, _ *emptypb.Empty) (*v0Status.BuildInfo, error) {
	ctx, span := g.tracer.Start(ctx, "status.GrpcHandler.GetVersion")
	defer span.End()

	return v0StatusBuildInfo(), nil
}

func v0StatusBuildInfo() *v0Status.BuildInfo {
	var ret *v0Status.BuildInfo = nil

	if info := buildInfo(); info != nil {
		ret = &v0Status.BuildInfo{
			Version:    info.Version,
			CommitHash: info.CommitHash,
			Name:       info.Name,
		}
	}

	return ret
}

func NewGrpcHandler(tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GrpcHandler {
	g := new(GrpcHandler)

	g.tracer = tracer
	g.monitor = monitor
	g.logger = logger

	return g
}
