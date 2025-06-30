// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"context"
	"io/fs"
	"net/http"

	v0Clients "github.com/canonical/identity-platform-api/v0/clients"

	v0Groups "github.com/canonical/identity-platform-api/v0/groups"
	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
	v0Roles "github.com/canonical/identity-platform-api/v0/roles"
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	v0Status "github.com/canonical/identity-platform-api/v0/status"
	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
	"github.com/canonical/identity-platform-admin-ui/pkg/entitlements"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/metrics"
	"github.com/canonical/identity-platform-admin-ui/pkg/resources"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/status"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"

	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
)

type RouterOption func(*routerConfig)

func WithIDPConfig(cfg *idp.Config) RouterOption {
	return func(c *routerConfig) {
		c.idp = cfg
	}
}

func WithSchemasConfig(cfg *schemas.Config) RouterOption {
	return func(c *routerConfig) {
		c.schemas = cfg
	}
}

func WithUIDistFS(distFS *fs.FS) RouterOption {
	return func(c *routerConfig) {
		c.uiDistFS = distFS
	}
}

func WithExternalClients(cfg ExternalClientsConfigInterface) RouterOption {
	return func(c *routerConfig) {
		c.external = cfg
	}
}

func WithOAuth2Config(cfg *authentication.Config) RouterOption {
	return func(c *routerConfig) {
		c.oauth2 = cfg
	}
}

func WithMailConfig(cfg *mail.Config) RouterOption {
	return func(c *routerConfig) {
		c.mail = cfg
	}
}

func WithO11y(cfg O11yConfigInterface) RouterOption {
	return func(c *routerConfig) {
		c.olly = cfg
	}
}

func WithContextPath(path string) RouterOption {
	return func(c *routerConfig) {
		c.contextPath = path
	}
}

func WithPayloadValidationEnabled(enabled bool) RouterOption {
	return func(c *routerConfig) {
		c.payloadValidationEnabled = enabled
	}
}

type routerConfig struct {
	contextPath              string
	payloadValidationEnabled bool
	idp                      *idp.Config
	schemas                  *schemas.Config
	uiDistFS                 *fs.FS
	external                 ExternalClientsConfigInterface
	oauth2                   *authentication.Config
	mail                     *mail.Config
	olly                     O11yConfigInterface
}

func NewRouter(wpool pool.WorkerPoolInterface, dbClient storage.DBClientInterface, opts ...RouterOption) http.Handler {
	router := chi.NewMux()

	config := &routerConfig{}
	for _, opt := range opts {
		opt(config)
	}

	externalConfig := config.external
	oauth2Config := config.oauth2

	logger := config.olly.Logger()
	monitor := config.olly.Monitor()
	tracer := config.olly.Tracer()
	store := ofga.NewOpenFGAStore(externalConfig.OpenFGA(), wpool, tracer, monitor, logger)

	middlewares := make(chi.Middlewares, 0)
	middlewares = append(
		middlewares,
		middleware.RequestID,
		monitoring.NewMiddleware(monitor, logger).ResponseTime(),
		middlewareCORS([]string{"*"}),
	)
	authorizationMiddleware := authorization.NewMiddleware(config.external.Authorizer(), monitor, logger).Authorize()

	// TODO @shipperizer add a proper configuration to enable http logger middleware as it's expensive
	if true {
		middlewares = append(
			middlewares,
			middleware.RequestLogger(logging.NewLogFormatter(logger)), // LogFormatter will only work if logger is set to DEBUG level
		)
	}

	roleRepository := roles.NewRoleRepository(dbClient, tracer, monitor, logger)

	mailService := mail.NewEmailService(config.mail, tracer, monitor, logger)

	identitiesSvc := identities.NewService(externalConfig.KratosAdmin().IdentityAPI(), externalConfig.Authorizer(), mailService, tracer, monitor, logger)
	idpSvc := idp.NewService(config.idp, externalConfig.Authorizer(), tracer, monitor, logger)
	rolesSvc := roles.NewService(externalConfig.OpenFGA(), roleRepository, wpool, tracer, monitor, logger)
	groupsSvc := groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger)
	schemaSvc := schemas.NewService(config.schemas, externalConfig.Authorizer(), tracer, monitor, logger)
	clientsSvc := clients.NewService(externalConfig.HydraAdmin(), externalConfig.Authorizer(), tracer, monitor, logger)

	router.Use(middlewares...)

	identitiesAPI := identities.NewAPI(
		identitiesSvc,
		tracer,
		monitor,
		logger,
	)

	clientsAPI := clients.NewAPI(
		clientsSvc,
		tracer,
		monitor,
		logger,
	)

	idpAPI := idp.NewAPI(
		idpSvc,
		tracer,
		monitor,
		logger,
	)

	schemasAPI := schemas.NewAPI(
		schemaSvc,
		tracer,
		monitor,
		logger,
	)

	rolesAPI := roles.NewAPI(
		rolesSvc,
		tracer,
		monitor,
		logger,
	)

	groupsAPI := groups.NewAPI(
		groupsSvc,
		tracer,
		monitor,
		logger,
	)

	// Create a new router for the API so that we can add extra middlewares
	apiRouter := router.Group(nil).(*chi.Mux)

	var oauth2Context authentication.OAuth2ContextInterface
	var cookieManager authentication.AuthCookieManagerInterface

	if oauth2Config.Enabled {
		oauth2Context = authentication.NewOAuth2Context(config.oauth2, oidc.NewProvider, tracer, logger, monitor)
		encrypt := authentication.NewEncrypt([]byte(oauth2Config.CookiesEncryptionKey), logger, tracer)
		cookieManager = authentication.NewAuthCookieManager(
			oauth2Config.AuthCookieTTLSeconds,
			oauth2Config.UserSessionCookieTTLSeconds,
			encrypt,
			logger,
		)

		authenticationMiddleware := authentication.NewAuthenticationMiddleware(oauth2Context, cookieManager, tracer, logger)
		authenticationMiddleware.SetAllowListedEndpoints(
			"/api/v0/auth",
			"/api/v0/auth/callback",
			"/api/v0/status",
			"/api/v0/metrics",
		)
		apiRouter.Use(authenticationMiddleware.OAuth2AuthenticationChain()...)
	} else {
		apiRouter.Use(authentication.AuthenticationDisabledMiddleware)
	}

	// register authorizationMiddleware after authentication so Principal is available if necessary
	apiRouter.Use(authorizationMiddleware)

	if config.payloadValidationEnabled {
		validationRegistry := validation.NewRegistry(tracer, monitor, logger)
		apiRouter.Use(validationRegistry.ValidationMiddleware)

		identitiesAPI.RegisterValidation(validationRegistry)
		clientsAPI.RegisterValidation(validationRegistry)
		idpAPI.RegisterValidation(validationRegistry)
		schemasAPI.RegisterValidation(validationRegistry)
		rolesAPI.RegisterValidation(validationRegistry)
		groupsAPI.RegisterValidation(validationRegistry)
	}

	// register endpoints as last step
	//statusAPI.RegisterEndpoints(apiRouter)
	//metricsAPI.RegisterEndpoints(apiRouter)

	//identitiesAPI.RegisterEndpoints(apiRouter)
	//clientsAPI.RegisterEndpoints(apiRouter)
	//idpAPI.RegisterEndpoints(apiRouter)
	//schemasAPI.RegisterEndpoints(apiRouter)
	// while we port APIs to the new gRPC-gateway based implementation, we disable the original ones step by step
	// rolesAPI.RegisterEndpoints(apiRouter)
	// groupsAPI.RegisterEndpoints(apiRouter)

	/********* gRPC gateway integration **********/
	gRPCGatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(types.SetHeaderFromMetadataFilter),
		runtime.WithForwardResponseRewriter(types.ForwardErrorResponseRewriter),
	)

	err := v0Roles.RegisterRolesServiceHandlerServer(context.Background(), gRPCGatewayMux, roles.NewGrpcHandler(rolesSvc, tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Groups.RegisterGroupsServiceHandlerServer(context.Background(), gRPCGatewayMux, groups.NewGrpcHandler(groupsSvc, tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Status.RegisterStatusServiceHandlerServer(context.Background(), gRPCGatewayMux, status.NewGrpcHandler(tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Identities.RegisterIdentitiesServiceHandlerServer(context.Background(), gRPCGatewayMux, identities.NewGrpcHandler(identitiesSvc, identities.NewGrpcMapper(logger), tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Idps.RegisterIdpsServiceHandlerServer(context.Background(), gRPCGatewayMux, idp.NewGrpcHandler(idpSvc, idp.NewGrpcMapper(logger), tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Schemas.RegisterSchemasServiceHandlerServer(context.Background(), gRPCGatewayMux, schemas.NewGrpcHandler(schemaSvc, schemas.NewGrpcMapper(logger), tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	err = v0Clients.RegisterClientsServiceHandlerServer(context.Background(), gRPCGatewayMux, clients.NewGrpcHandler(clientsSvc, clients.NewGrpcMapper(logger), tracer, monitor, logger))
	if err != nil {
		panic(err)
	}

	apiRouter.Mount("/api/v0/identities", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/roles", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/groups", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/idps", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/schemas", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/clients", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/status", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/version", gRPCGatewayMux)
	apiRouter.Mount("/api/v0/metrics", metrics.NewAPI(logger).Handler())
	/********* gRPC gateway integration **********/

	if oauth2Config.Enabled {

		sessionManager := authentication.NewSessionManagerService(externalConfig.KratosAdmin().IdentityAPI(), externalConfig.KratosPublic().FrontendAPI(), tracer, monitor, logger)
		login := authentication.NewAPI(
			config.contextPath,
			oauth2Context,
			authentication.NewOAuth2Helper(),
			cookieManager,
			sessionManager,
			tracer,
			logger,
		)
		login.RegisterEndpoints(apiRouter)
	}

	rebacAPI, err := v1.NewReBACAdminBackend(
		v1.ReBACAdminBackendParams{
			Resources: resources.NewV1Service(store, tracer, monitor, logger),
			Roles:     roles.NewV1Service(rolesSvc),
			Groups:    groups.NewV1Service(groupsSvc, tracer, monitor, logger),
			Identities: identities.NewV1Service(
				&identities.Config{
					Name:         config.idp.Name,
					Namespace:    config.idp.Namespace,
					K8s:          config.idp.K8s,
					OpenFGAStore: store,
				},
				identitiesSvc,
			),
			Entitlements:      entitlements.NewV1Service(externalConfig.OpenFGA(), tracer, monitor, logger),
			IdentityProviders: idp.NewV1Service(idpSvc),
		},
	)

	if err != nil {
		panic(err)
	}

	apiRouter.Mount("/api/", rebacAPI.Handler(""))

	ui.NewAPI(config.contextPath, *config.uiDistFS, tracer, monitor, logger).
		RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
