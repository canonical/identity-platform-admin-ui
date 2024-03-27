// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package web

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
	trace "go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	ih "github.com/canonical/identity-platform-admin-ui/internal/hydra"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"

	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/metrics"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
	"github.com/canonical/identity-platform-admin-ui/pkg/rules"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/status"
)

func NewRouter(idpConfig *idp.Config, schemasConfig *schemas.Config, rulesConfig *rules.Config, hydraClient *ih.Client, kratos *ik.Client, ofga OpenFGAClientInterface, authorizationClient OpenFGAClientInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) http.Handler {
	router := chi.NewMux()

	middlewares := make(chi.Middlewares, 0)
	middlewares = append(
		middlewares,
		middleware.RequestID,
		monitoring.NewMiddleware(monitor, logger).ResponseTime(),
		middlewareCORS([]string{"*"}),
	)

	// TODO @shipperizer add a proper configuration to enable http logger middleware as it's expensive
	if true {
		middlewares = append(
			middlewares,
			middleware.RequestLogger(logging.NewLogFormatter(logger)), // LogFormatter will only work if logger is set to DEBUG level
		)
	}

	router.Use(middlewares...)

	// apply authorization middleware using With to overcome issue with <id> URLParams not available
	router = router.With(
		authorization.NewMiddleware(
			authorization.NewAuthorizer(authorizationClient, tracer, monitor, logger), monitor, logger).Authorize(),
	).(*chi.Mux)

	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)
	identities.NewAPI(
		identities.NewService(kratos.IdentityApi(), tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	clients.NewAPI(
		clients.NewService(hydraClient, tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	idp.NewAPI(
		idp.NewService(idpConfig, tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	schemas.NewAPI(
		schemas.NewService(schemasConfig, tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	rules.NewAPI(
		rules.NewService(rulesConfig, tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	roles.NewAPI(
		roles.NewService(ofga, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	).RegisterEndpoints(router)
	groups.NewAPI(
		groups.NewService(ofga, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	).RegisterEndpoints(router)
	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
