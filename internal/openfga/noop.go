// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package openfga

import (
	"context"

	openfga "github.com/openfga/go-sdk"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"

	"github.com/canonical/identity-platform-admin-ui/internal/tracing"

	"github.com/openfga/go-sdk/client"
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

func (c *NoopClient) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	return make([]string, 0), nil
}

func (c *NoopClient) Check(ctx context.Context, user, relation, object string) (bool, error) {
	return true, nil
}

func (c *NoopClient) BatchCheck(ctx context.Context, tuples ...Tuple) (bool, error) {
	return true, nil
}

func (c *NoopClient) WriteTuple(ctx context.Context, user, relation, object string) error {
	return nil
}

func (c *NoopClient) WriteTuples(ctx context.Context, tuples ...Tuple) error {
	return nil
}

func (c *NoopClient) DeleteTuple(ctx context.Context, user, relation, object string) error {
	return nil
}

func (c *NoopClient) DeleteTuples(ctx context.Context, tuples ...Tuple) error {
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

func (c *NoopClient) ReadTuples(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
	return nil, nil
}
