// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

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
	"github.com/canonical/identity-platform-admin-ui/internal/validation"

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

	vldtr := validation.NewRegistry(tracer, monitor, logger)

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

	// apply authorization and validation middlewares using With to overcome issue with <id> URLParams not available
	router = router.With(
		authorization.NewMiddleware(
			authorization.NewAuthorizer(externalConfig.Authorizer(), tracer, monitor, logger), monitor, logger).Authorize(),
		vldtr.ValidationMiddleware,
	).(*chi.Mux)

	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)

	identitiesAPI := identities.NewAPI(
		identities.NewService(externalConfig.KratosAdmin().IdentityApi(), tracer, monitor, logger),
		logger,
	)
	identitiesAPI.RegisterEndpoints(router)
	identitiesAPI.RegisterValidation(vldtr)

	clientsAPI := clients.NewAPI(
		clients.NewService(externalConfig.HydraAdmin(), tracer, monitor, logger),
		logger,
	)
	clientsAPI.RegisterEndpoints(router)
	clientsAPI.RegisterValidation(vldtr)

	idpAPI := idp.NewAPI(
		idp.NewService(idpConfig, tracer, monitor, logger),
		logger,
	)
	idpAPI.RegisterEndpoints(router)
	idpAPI.RegisterValidation(vldtr)

	schemasAPI := schemas.NewAPI(
		schemas.NewService(schemasConfig, tracer, monitor, logger),
		logger,
	)
	schemasAPI.RegisterEndpoints(router)
	schemasAPI.RegisterValidation(vldtr)

	rulesAPI := rules.NewAPI(
		rules.NewService(rulesConfig, tracer, monitor, logger),
		logger,
	)
	rulesAPI.RegisterEndpoints(router)
	rulesAPI.RegisterValidation(vldtr)

	rolesAPI := roles.NewAPI(
		roles.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)
	rolesAPI.RegisterEndpoints(router)
	rolesAPI.RegisterValidation(vldtr)

	groupsAPI := groups.NewAPI(
		groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)
	groupsAPI.RegisterEndpoints(router)
	groupsAPI.RegisterValidation(vldtr)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
