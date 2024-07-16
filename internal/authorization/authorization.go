// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package authorization

import (
	"context"
	"fmt"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

var ErrInvalidAuthModel = fmt.Errorf("Invalid authorization model schema")

type Authorizer struct {
	client AuthzClientInterface

	wpool pool.WorkerPoolInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface

	AdminAuthorizer
}

func (a *Authorizer) Check(ctx context.Context, user string, relation string, object string, contextualTuples ...openfga.Tuple) (bool, error) {
	ctx, span := a.tracer.Start(ctx, "authorization.Authorizer.Check")
	defer span.End()

	return a.client.Check(ctx, user, relation, object, contextualTuples...)
}

func (a *Authorizer) ListObjects(ctx context.Context, user string, relation string, objectType string) ([]string, error) {
	ctx, span := a.tracer.Start(ctx, "authorization.Authorizer.ListObjects")
	defer span.End()

	return a.client.ListObjects(ctx, user, relation, objectType)
}

func (a *Authorizer) FilterObjects(ctx context.Context, user string, relation string, objectType string, objs []string) ([]string, error) {
	ctx, span := a.tracer.Start(ctx, "authorization.Authorizer.FilterObjects")
	defer span.End()

	allowedObjs, err := a.ListObjects(ctx, user, relation, objectType)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, obj := range allowedObjs {
		if contains(objs, obj) {
			ret = append(ret, obj)
		}
	}
	return ret, nil
}

func (a *Authorizer) ValidateModel(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "authorization.Authorizer.ValidateModel")
	defer span.End()

	eq, err := a.client.CompareModel(ctx, AuthModel)
	if err != nil {
		return err
	}
	if !eq {
		return ErrInvalidAuthModel
	}
	return nil
}

func NewAuthorizer(client AuthzClientInterface, wpool pool.WorkerPoolInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Authorizer {
	authorizer := new(Authorizer)
	authorizer.client = client
	authorizer.wpool = wpool
	authorizer.tracer = tracer
	authorizer.monitor = monitor
	authorizer.logger = logger
	authorizer.AdminAuthorizer = *NewAdminAuthorizer(client, tracer, monitor, logger)

	return authorizer
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
