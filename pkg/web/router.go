// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package web

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validator"

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

func NewRouter(idpConfig *idp.Config, schemasConfig *schemas.Config, rulesConfig *rules.Config, externalConfig ExternalClientsConfigInterface, wpool pool.WorkerPoolInterface, ollyConfig O11yConfigInterface) http.Handler {
	router := chi.NewMux()

	logger := ollyConfig.Logger()
	monitor := ollyConfig.Monitor()
	tracer := ollyConfig.Tracer()

	vldtr := validation.NewValidator(tracer, monitor, logger)

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
			authorization.NewAuthorizer(externalConfig.Authorizer(), tracer, monitor, logger), monitor, logger).Authorize(),
	).(*chi.Mux)

	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)
	identities.NewAPI(
		identities.NewService(externalConfig.KratosAdmin().IdentityApi(), tracer, monitor, logger),
		logger,
	).RegisterEndpoints(router)
	clients.NewAPI(
		clients.NewService(externalConfig.HydraAdmin(), tracer, monitor, logger),
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
		roles.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	).RegisterEndpoints(router)
	groups.NewAPI(
		groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	).RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
