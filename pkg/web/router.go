// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"context"
	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	"net/http"

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
	"github.com/canonical/identity-platform-admin-ui/pkg/rules"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/status"

	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
)

type RouterConfig struct {
	contextPath              string
	payloadValidationEnabled bool
	idp                      *idp.Config
	schemas                  *schemas.Config
	rules                    *rules.Config
	ui                       *ui.Config
	external                 ExternalClientsConfigInterface
	oauth2                   *authentication.Config
	mail                     *mail.Config
	olly                     O11yConfigInterface
}

func NewRouterConfig(contextPath string, payloadValidationEnabled bool, idp *idp.Config, schemas *schemas.Config, rules *rules.Config, ui *ui.Config, external ExternalClientsConfigInterface, oauth2 *authentication.Config, mail *mail.Config, olly O11yConfigInterface) *RouterConfig {
	return &RouterConfig{
		contextPath:              contextPath,
		payloadValidationEnabled: payloadValidationEnabled,
		idp:                      idp,
		schemas:                  schemas,
		rules:                    rules,
		ui:                       ui,
		external:                 external,
		oauth2:                   oauth2,
		mail:                     mail,
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
	mailConfig := config.mail

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

	mailService := mail.NewEmailService(mailConfig, tracer, monitor, logger)

	identitiesSvc := identities.NewService(externalConfig.KratosAdmin().IdentityAPI(), externalConfig.Authorizer(), mailService, tracer, monitor, logger)
	idpSvc := idp.NewService(idpConfig, externalConfig.Authorizer(), tracer, monitor, logger)
	rolesSvc := roles.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger)
	groupsSvc := groups.NewService(externalConfig.OpenFGA(), wpool, tracer, monitor, logger)
	schemaSvc := schemas.NewService(schemasConfig, externalConfig.Authorizer(), tracer, monitor, logger)
	clientsSvc := clients.NewService(externalConfig.HydraAdmin(), externalConfig.Authorizer(), tracer, monitor, logger)

	router.Use(middlewares...)

	//statusAPI := status.NewAPI(tracer, monitor, logger)
	metricsAPI := metrics.NewAPI(logger)

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

	rulesAPI := rules.NewAPI(
		rules.NewService(rulesConfig, externalConfig.Authorizer(), tracer, monitor, logger),
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

	uiAPI := ui.NewAPI(uiConfig, tracer, monitor, logger)

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
		rulesAPI.RegisterValidation(validationRegistry)
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
	rulesAPI.RegisterEndpoints(apiRouter)
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
	apiRouter.Mount("/api/v0/metrics", metricsAPI.Handler())
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
					Name:         idpConfig.Name,
					Namespace:    idpConfig.Namespace,
					K8s:          idpConfig.K8s,
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

	uiAPI.RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
