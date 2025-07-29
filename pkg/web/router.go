// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package web

import (
	"io/fs"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/config"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
	"github.com/canonical/identity-platform-admin-ui/pkg/web/v0"
	"github.com/canonical/identity-platform-admin-ui/pkg/web/v1"
	v2 "github.com/canonical/identity-platform-admin-ui/pkg/web/v2"

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

func WithExternalClients(cfg config.ExternalClientsConfigInterface) RouterOption {
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

func WithO11y(cfg config.O11yConfigInterface) RouterOption {
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
	external                 config.ExternalClientsConfigInterface
	oauth2                   *authentication.Config
	mail                     *mail.Config
	olly                     config.O11yConfigInterface
}

func NewRouter(wpool pool.WorkerPoolInterface, dbClient storage.DBClientInterface, opts ...RouterOption) http.Handler {
	router := chi.NewMux()

	cfg := &routerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	externalConfig := cfg.external

	logger := cfg.olly.Logger()
	monitor := cfg.olly.Monitor()
	tracer := cfg.olly.Tracer()
	store := ofga.NewOpenFGAStore(externalConfig.OpenFGA(), wpool, tracer, monitor, logger)

	middlewares := make(chi.Middlewares, 0)
	middlewares = append(
		middlewares,
		middleware.RequestID,
		monitoring.NewMiddleware(monitor, logger).ResponseTime(),
		middlewareCORS([]string{"*"}),
	)
	authorizationMiddleware := authorization.NewMiddleware(cfg.external.Authorizer(), monitor, logger).Authorize()

	// TODO @shipperizer add a proper configuration to enable http logger middleware as it's expensive
	if true {
		middlewares = append(
			middlewares,
			middleware.RequestLogger(logging.NewLogFormatter(logger)), // LogFormatter will only work if logger is set to DEBUG level
		)
	}

	roleRepository := roles.NewRoleRepository(dbClient, tracer, monitor, logger)
	groupRepository := groups.NewGroupRepository(dbClient, tracer, monitor, logger)

	mailService := mail.NewEmailService(cfg.mail, tracer, monitor, logger)

	identitiesSvc := identities.NewService(externalConfig.KratosAdmin().IdentityAPI(), externalConfig.Authorizer(), mailService, tracer, monitor, logger)
	idpSvc := idp.NewService(cfg.idp, externalConfig.Authorizer(), tracer, monitor, logger)
	rolesSvc := roles.NewService(externalConfig.OpenFGA(), roleRepository, wpool, tracer, monitor, logger)
	groupsSvc := groups.NewService(externalConfig.OpenFGA(), groupRepository, wpool, tracer, monitor, logger)
	schemaSvc := schemas.NewService(cfg.schemas, externalConfig.Authorizer(), tracer, monitor, logger)
	clientsSvc := clients.NewService(externalConfig.HydraAdmin(), externalConfig.Authorizer(), tracer, monitor, logger)

	router.Use(middlewares...)

	// Create a new router for the API so that we can add extra middlewares
	apiRouter := router.Group(nil).(*chi.Mux)

	var oauth2Context authentication.OAuth2ContextInterface
	var cookieManager authentication.AuthCookieManagerInterface

	if cfg.oauth2.Enabled {
		oauth2Context = authentication.NewOAuth2Context(
			cfg.oauth2,
			oidc.NewProvider,
			tracer,
			logger,
			monitor,
		)
		encrypt := authentication.NewEncrypt(
			[]byte(cfg.oauth2.CookiesEncryptionKey),
			logger,
			tracer,
		)

		cookieManager = authentication.NewAuthCookieManager(
			cfg.oauth2.AuthCookieTTLSeconds,
			cfg.oauth2.UserSessionCookieTTLSeconds,
			encrypt,
			logger,
		)

		authenticationMiddleware := authentication.NewAuthenticationMiddleware(
			oauth2Context,
			cookieManager,
			tracer,
			logger,
		)
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

	v0Api, err := v0.NewV0APIRouter(
		v0.WithContextPath(cfg.contextPath),
		v0.WithOAuth2Context(oauth2Context),
		v0.WithCookieManager(cookieManager),
		v0.WithPayloadValidation(cfg.payloadValidationEnabled),
		v0.WithIDPConfig(cfg.idp),
		v0.WithStore(store),
		v0.WithRolesService(rolesSvc),
		v0.WithGroupsService(groupsSvc),
		v0.WithClientsService(clientsSvc),
		v0.WithIdentitiesService(identitiesSvc),
		v0.WithIDPService(idpSvc),
		v0.WithSchemasService(schemaSvc),
		v0.WithOpenFGA(externalConfig.OpenFGA()),
		v0.WithKratos(externalConfig.KratosAdmin(), externalConfig.KratosPublic()),
		v0.WithO11yConfig(cfg.olly),
	)
	if err != nil {
		panic(err)
	}

	v1Api, err := v1.NewV1APIRouter(
		v1.WithIDPConfig(cfg.idp),
		v1.WithStore(store),
		v1.WithRolesService(rolesSvc),
		v1.WithGroupsService(groupsSvc),
		v1.WithIdentitiesService(identitiesSvc),
		v1.WithIDPService(idpSvc),
		v1.WithOpenFGA(externalConfig.OpenFGA()),
		v1.WithO11yConfig(cfg.olly),
	)

	if err != nil {
		panic(err)
	}

	if cfg.oauth2.Enabled {
		authentication.NewAPI(
			cfg.contextPath,
			oauth2Context,
			authentication.NewOAuth2Helper(),
			cookieManager,
			authentication.NewSessionManagerService(
				externalConfig.KratosAdmin().IdentityAPI(),
				externalConfig.KratosPublic().FrontendAPI(),
				tracer,
				monitor,
				logger,
			),
			tracer,
			logger,
		).RegisterEndpoints(apiRouter)
	}

	v2Api, err := v2.NewV2APIRouter()
	if err != nil {
		panic(err)
	}

	apiRouter.Mount("/api/v0", v0Api)
	apiRouter.Mount("/api", v1Api) // v1 library provides the `/v1` prefix already
	apiRouter.Mount("/api/v2", v2Api)

	ui.NewAPI(cfg.contextPath, *cfg.uiDistFS, tracer, monitor, logger).
		RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
