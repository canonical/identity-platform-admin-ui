// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/metrics"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
	"github.com/canonical/identity-platform-admin-ui/pkg/rules"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/status"
	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
)

type RouterConfig struct {
	payloadValidationEnabled bool
	idp                      *idp.Config
	schemas                  *schemas.Config
	rules                    *rules.Config
	ui                       *ui.Config
	external                 ExternalClientsConfigInterface
	oauth2                   *authentication.Config
	olly                     O11yConfigInterface
}

func NewRouterConfig(payloadValidationEnabled bool, idp *idp.Config, schemas *schemas.Config, rules *rules.Config, ui *ui.Config, external ExternalClientsConfigInterface, oauth2 *authentication.Config, olly O11yConfigInterface) *RouterConfig {
	return &RouterConfig{
		payloadValidationEnabled: payloadValidationEnabled,
		idp:                      idp,
		schemas:                  schemas,
		rules:                    rules,
		ui:                       ui,
		external:                 external,
		oauth2:                   oauth2,
		olly:                     olly,
	}
}

func NewRouter(config *RouterConfig, wpool pool.WorkerPoolInterface) http.Handler {
	router := chi.NewMux()

	idpConfig := config.idp
	schemasConfig := config.schemas
	rulesConfig := config.rules
	uiConfig := config.ui
	externalConfig := config.external
	oauth2Config := config.oauth2

	logger := config.olly.Logger()
	monitor := config.olly.Monitor()
	tracer := config.olly.Tracer()

	middlewares := make(chi.Middlewares, 0)
	middlewares = append(
		middlewares,
		middleware.RequestID,
		monitoring.NewMiddleware(monitor, logger).ResponseTime(),
		middlewareCORS([]string{"*"}),
	)
	authorizationMiddleware := authorization.NewMiddleware(
		authorization.NewAuthorizer(externalConfig.Authorizer(), tracer, monitor, logger), monitor, logger,
	).Authorize()

	// TODO @shipperizer add a proper configuration to enable http logger middleware as it's expensive
	if true {
		middlewares = append(
			middlewares,
			middleware.RequestLogger(logging.NewLogFormatter(logger)), // LogFormatter will only work if logger is set to DEBUG level
		)
	}

	router.Use(middlewares...)

	statusAPI := status.NewAPI(tracer, monitor, logger)
	metricsAPI := metrics.NewAPI(logger)

	identitiesAPI := identities.NewAPI(
		identities.NewService(externalConfig.KratosAdmin().IdentityAPI(), tracer, monitor, logger),
		logger,
	)

	clientsAPI := clients.NewAPI(
		clients.NewService(externalConfig.HydraAdmin(), tracer, monitor, logger),
		logger,
	)

	idpAPI := idp.NewAPI(
		idp.NewService(idpConfig, tracer, monitor, logger),
		logger,
	)

	schemasAPI := schemas.NewAPI(
		schemas.NewService(schemasConfig, tracer, monitor, logger),
		logger,
	)

	rulesAPI := rules.NewAPI(
		rules.NewService(rulesConfig, tracer, monitor, logger),
		logger,
	)

	rolesAPI := roles.NewAPI(
		roles.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)

	groupsAPI := groups.NewAPI(
		groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger),
		tracer,
		monitor,
		logger,
	)

	uiAPI := ui.NewAPI(uiConfig, tracer, monitor, logger)

	// Create a new router for the API so that we can add extra middlewares
	apiRouter := router.Group(func(r chi.Router) {
		r.Use(authorizationMiddleware)
	}).(*chi.Mux)

	var oauth2Context *authentication.OAuth2Context

	if oauth2Config.Enabled {
		oauth2Context = authentication.NewOAuth2Context(config.oauth2, oidc.NewProvider, tracer, logger, monitor)

		authenticationMiddleware := authentication.NewAuthenticationMiddleware(oauth2Context, tracer, logger)
		authenticationMiddleware.SetAllowListedEndpoints(
			"/api/v0/auth",
			"/api/v0/auth/callback",
			"/api/v0/status",
			"/api/v0/metrics",
		)
		apiRouter.Use(authenticationMiddleware.OAuth2Authentication)
	}

	if config.payloadValidationEnabled {
		validationRegistry := validation.NewRegistry(tracer, monitor, logger)
		apiRouter.Use(validationRegistry.ValidationMiddleware)

		identitiesAPI.RegisterValidation(validationRegistry)
		clientsAPI.RegisterValidation(validationRegistry)
		idpAPI.RegisterValidation(validationRegistry)
		schemasAPI.RegisterValidation(validationRegistry)
		rulesAPI.RegisterValidation(validationRegistry)
		rolesAPI.RegisterValidation(validationRegistry)
		groupsAPI.RegisterValidation(validationRegistry)
	}

	// register endpoints as last step
	statusAPI.RegisterEndpoints(apiRouter)
	metricsAPI.RegisterEndpoints(apiRouter)

	identitiesAPI.RegisterEndpoints(apiRouter)
	clientsAPI.RegisterEndpoints(apiRouter)
	idpAPI.RegisterEndpoints(apiRouter)
	schemasAPI.RegisterEndpoints(apiRouter)
	rulesAPI.RegisterEndpoints(apiRouter)
	rolesAPI.RegisterEndpoints(apiRouter)
	groupsAPI.RegisterEndpoints(apiRouter)

	if oauth2Config.Enabled {
		login := authentication.NewAPI(oauth2Config.AuthCookieTTLSeconds, oauth2Context, authentication.NewOAuth2Helper(), authentication.NewAuthCookieManager(), tracer, logger)
		login.RegisterEndpoints(apiRouter)
	}

	uiAPI.RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
