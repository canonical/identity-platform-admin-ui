// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	ih "github.com/canonical/identity-platform-admin-ui/internal/hydra"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

type AuthorizerClientInterface = *authorization.Authorizer

type O11yConfigInterface interface {
	Tracer() trace.Tracer
	Monitor() monitoring.MonitorInterface
	Logger() logging.LoggerInterface
}

type ExternalClientsConfigInterface interface {
	HydraAdmin() *ih.Client
	KratosAdmin() *ik.Client
	KratosPublic() *ik.Client
	OpenFGA() openfga.OpenFGAClientInterface
	Authorizer() AuthorizerClientInterface
}
