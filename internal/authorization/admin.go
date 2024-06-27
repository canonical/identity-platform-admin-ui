// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package authorization

import (
	"context"
	"fmt"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const PRIVILEGED_RELATION = "privileged"
const ADMIN_OBJECT = "privileged:superuser"

type AdminAuthorizer struct {
	client AuthzClientInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (a *AdminAuthorizer) CreateAdmin(ctx context.Context, username string) error {
	ctx, span := a.tracer.Start(ctx, "authorization.AdminAuthorizer.CreateAdmin")
	defer span.End()

	user := fmt.Sprintf("user:%s", username)
	err := a.client.WriteTuple(ctx, user, "admin", ADMIN_OBJECT)
	return err
}

func (a *AdminAuthorizer) RemoveAdmin(ctx context.Context, username string) error {
	ctx, span := a.tracer.Start(ctx, "authorization.AdminAuthorizer.RemoveAdmin")
	defer span.End()

	user := fmt.Sprintf("user:%s", username)
	err := a.client.DeleteTuple(ctx, user, "admin", ADMIN_OBJECT)
	return err
}

func (a *AdminAuthorizer) CheckAdmin(ctx context.Context, username string) (bool, error) {
	ctx, span := a.tracer.Start(ctx, "authorization.AdminAuthorizer.CheckAdmin")
	defer span.End()

	user := fmt.Sprintf("user:%s", username)
	allowed, err := a.client.Check(ctx, user, "admin", ADMIN_OBJECT)

	return allowed, err
}

func NewAdminAuthorizer(client AuthzClientInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *AdminAuthorizer {
	authorizer := new(AdminAuthorizer)
	authorizer.client = client
	authorizer.tracer = tracer
	authorizer.monitor = monitor
	authorizer.logger = logger

	return authorizer
}
