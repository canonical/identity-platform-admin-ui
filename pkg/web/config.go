// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	trace "go.opentelemetry.io/otel/trace"

	ih "github.com/canonical/identity-platform-admin-ui/internal/hydra"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

// O11yConfig is a wrapper config for all the observability objects
type O11yConfig struct {
	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// Tracer returns the tracing object
func (c *O11yConfig) Tracer() trace.Tracer {
	return c.tracer
}

// Monitor returns a monitor object
func (c *O11yConfig) Monitor() monitoring.MonitorInterface {
	return c.monitor
}

// Logger returns a logger object
func (c *O11yConfig) Logger() logging.LoggerInterface {
	return c.logger
}

// NewO11yConfig create an observability config object with a monitor, logger and tracer
func NewO11yConfig(tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *O11yConfig {
	c := new(O11yConfig)

	c.tracer = tracer
	c.monitor = monitor
	c.logger = logger

	return c
}

// ExternalClientsConfig is a wrapper config for all the third party clients
type ExternalClientsConfig struct {
	hydraAdmin   *ih.Client
	kratosAdmin  *ik.Client
	kratosPublic *ik.Client
	ofga         OpenFGAClientInterface
	authorizer   AuthorizerClientInterface
}

// HydraAdmin returns an hydra client to interact with the admin API
func (c *ExternalClientsConfig) HydraAdmin() *ih.Client {
	return c.hydraAdmin
}

// KratosAdmin returns a kratos client to interact with the admin API
func (c *ExternalClientsConfig) KratosAdmin() *ik.Client {
	return c.kratosAdmin
}

// KratosPublic returns a kratos client to interact with the public API
func (c *ExternalClientsConfig) KratosPublic() *ik.Client {
	return c.kratosPublic
}

// OpenFGA returns an openfga client
func (c *ExternalClientsConfig) OpenFGA() OpenFGAClientInterface {
	return c.ofga
}

// Authorizer returns an openfga client used for the authorization middleware
func (c *ExternalClientsConfig) Authorizer() AuthorizerClientInterface {
	return c.authorizer
}

// SetAuthorizer sets the authorization middleware
func (c *ExternalClientsConfig) SetAuthorizer(o AuthorizerClientInterface) {
	c.authorizer = o
}

// NewExternalClientsConfig create a third party config object for all the external clients needed
func NewExternalClientsConfig(hydra *ih.Client, kratosAdmin *ik.Client, kratosPublic *ik.Client, ofga OpenFGAClientInterface, authorizer AuthorizerClientInterface) *ExternalClientsConfig {
	c := new(ExternalClientsConfig)

	c.hydraAdmin = hydra
	c.kratosAdmin = kratosAdmin
	c.kratosPublic = kratosPublic
	c.ofga = ofga
	c.authorizer = authorizer

	return c
}
