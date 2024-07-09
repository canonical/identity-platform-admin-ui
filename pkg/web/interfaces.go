// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"context"

	fga "github.com/openfga/go-sdk"
	openfga "github.com/openfga/go-sdk"
	trace "go.opentelemetry.io/otel/trace"

	ih "github.com/canonical/identity-platform-admin-ui/internal/hydra"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

type OpenFGAClientInterface interface {
	ReadModel(context.Context) (*fga.AuthorizationModel, error)
	CompareModel(context.Context, fga.AuthorizationModel) (bool, error)
	WriteTuple(context.Context, string, string, string) error
	DeleteTuple(context.Context, string, string, string) error
	Check(context.Context, string, string, string, ...ofga.Tuple) (bool, error)
	ListObjects(context.Context, string, string, string) ([]string, error)
	WriteTuples(context.Context, ...ofga.Tuple) error
	DeleteTuples(context.Context, ...ofga.Tuple) error
	BatchCheck(context.Context, ...ofga.Tuple) (bool, error)
	ReadTuples(context.Context, string, string, string, string) (*openfga.ReadResponse, error)
}

type AuthorizerClientInterface interface {
	ListObjects(context.Context, string, string, string) ([]string, error)
	Check(context.Context, string, string, string, ...ofga.Tuple) (bool, error)
}

type O11yConfigInterface interface {
	Tracer() trace.Tracer
	Monitor() monitoring.MonitorInterface
	Logger() logging.LoggerInterface
}

type ExternalClientsConfigInterface interface {
	HydraAdmin() *ih.Client
	KratosAdmin() *ik.Client
	OpenFGA() OpenFGAClientInterface
	Authorizer() AuthorizerClientInterface
}
