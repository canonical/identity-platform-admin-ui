// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

type RouterConfig struct {
	payloadValidationEnabled bool
	idp                      *idp.Config
	schemas                  *schemas.Config
	rules                    *rules.Config
	external                 ExternalClientsConfigInterface
	olly                     O11yConfigInterface
}

func NewRouterConfig(payloadValidationEnabled bool, idp *idp.Config, schemas *schemas.Config, rules *rules.Config, external ExternalClientsConfigInterface, olly O11yConfigInterface) *RouterConfig {
	return &RouterConfig{
		payloadValidationEnabled: payloadValidationEnabled,
		idp:                      idp,
		schemas:                  schemas,
		rules:                    rules,
		external:                 external,
		olly:                     olly,
	}
}

func NewRouter(config *RouterConfig, wpool pool.WorkerPoolInterface) http.Handler {
	router := chi.NewMux()

	idpConfig := config.idp
	schemasConfig := config.schemas
	rulesConfig := config.rules
	externalConfig := config.external

	logger := config.olly.Logger()
	monitor := config.olly.Monitor()
	tracer := config.olly.Tracer()

	validationRegistry := validation.NewRegistry(tracer, monitor, logger)

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
		validationRegistry.ValidationMiddleware,
	).(*chi.Mux)

	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)

	identitiesAPI := identities.NewAPI(
		identities.NewService(externalConfig.KratosAdmin().IdentityAPI(), tracer, monitor, logger),
		logger,
	)
	identitiesAPI.RegisterEndpoints(router)
	identitiesAPI.RegisterValidation(validationRegistry)

	clientsAPI := clients.NewAPI(
		clients.NewService(externalConfig.HydraAdmin(), tracer, monitor, logger),
		logger,
	)
	clientsAPI.RegisterEndpoints(router)
	clientsAPI.RegisterValidation(validationRegistry)

	idpAPI := idp.NewAPI(
		idp.NewService(idpConfig, tracer, monitor, logger),
		logger,
	)
	idpAPI.RegisterEndpoints(router)
	idpAPI.RegisterValidation(validationRegistry)

	schemasAPI := schemas.NewAPI(
		schemas.NewService(schemasConfig, tracer, monitor, logger),
		logger,
	)
	schemasAPI.RegisterEndpoints(router)
	schemasAPI.RegisterValidation(validationRegistry)

	rulesAPI := rules.NewAPI(
		rules.NewService(rulesConfig, tracer, monitor, logger),
		logger,
	)
	rulesAPI.RegisterEndpoints(router)
	rulesAPI.RegisterValidation(validationRegistry)

	rolesAPI := roles.NewAPI(
		roles.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)
	rolesAPI.RegisterEndpoints(router)
	rolesAPI.RegisterValidation(validationRegistry)

	groupsAPI := groups.NewAPI(
		groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)
	groupsAPI.RegisterEndpoints(router)
	groupsAPI.RegisterValidation(validationRegistry)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
