// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package openfga

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	openfga "github.com/openfga/go-sdk"
)

type NoopClient struct {
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func NewNoopClient(tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *NoopClient {
	c := new(NoopClient)
	c.tracer = tracer
	c.monitor = monitor
	c.logger = logger
	return c
}

func (c *NoopClient) ListObjects(ctx context.Context, user string, relation string, objectType string) ([]string, error) {
	return make([]string, 0), nil
}

func (c *NoopClient) Check(ctx context.Context, user string, relation string, object string) (bool, error) {
	return true, nil
}

func (c *NoopClient) WriteTuple(ctx context.Context, user string, relation string, object string) error {
	return nil
}

func (c *NoopClient) ReadModel(ctx context.Context) (*openfga.AuthorizationModel, error) {
	return nil, nil
}

func (c *NoopClient) WriteModel(ctx context.Context, model []byte) (string, error) {
	return "", nil
}

func (c *NoopClient) CompareModel(ctx context.Context, model openfga.AuthorizationModel) (bool, error) {
	return true, nil
}
