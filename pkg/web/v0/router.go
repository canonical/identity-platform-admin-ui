// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package v0

import (
	"context"
	"fmt"
	"net/http"

	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	v0Groups "github.com/canonical/identity-platform-api/v0/groups"
	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
	v0Roles "github.com/canonical/identity-platform-api/v0/roles"
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	v0Status "github.com/canonical/identity-platform-api/v0/status"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	kClient "github.com/ory/kratos-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/config"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/clients"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/metrics"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/status"
)

type RouterOption func(*v0RouterConfig)

type v0RouterConfig struct {
	contextPath              string
	payloadValidationEnabled bool

	oauth2Context     authentication.OAuth2ContextInterface
	cookieManager     authentication.AuthCookieManagerInterface
	idpConfig         *idp.Config
	store             *openfga.OpenFGAStore
	rolesSvc          *roles.Service
	groupsSvc         *groups.Service
	clientsSvc        *clients.Service
	identitiesSvc     *identities.Service
	idpSvc            *idp.Service
	schemaSvc         *schemas.Service
	openfga           openfga.OpenFGAClientInterface
	kratosIdentityAPI kClient.IdentityAPI
	kratosFrontendAPI kClient.FrontendAPI

	o11yConfig config.O11yConfigInterface
}

func WithContextPath(path string) RouterOption {
	return func(c *v0RouterConfig) {
		c.contextPath = path
	}
}

func WithOAuth2Context(oauth2Context authentication.OAuth2ContextInterface) RouterOption {
	return func(c *v0RouterConfig) {
		c.oauth2Context = oauth2Context
	}
}

func WithCookieManager(cookieManager authentication.AuthCookieManagerInterface) RouterOption {
	return func(c *v0RouterConfig) {
		c.cookieManager = cookieManager
	}
}

func WithPayloadValidation(enabled bool) RouterOption {
	return func(c *v0RouterConfig) {
		c.payloadValidationEnabled = enabled
	}
}

func WithIDPConfig(cfg *idp.Config) RouterOption {
	return func(c *v0RouterConfig) {
		c.idpConfig = cfg
	}
}

func WithStore(store *openfga.OpenFGAStore) RouterOption {
	return func(c *v0RouterConfig) {
		c.store = store
	}
}

func WithRolesService(svc *roles.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.rolesSvc = svc
	}
}

func WithGroupsService(svc *groups.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.groupsSvc = svc
	}
}

func WithClientsService(svc *clients.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.clientsSvc = svc
	}
}

func WithIdentitiesService(svc *identities.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.identitiesSvc = svc
	}
}

func WithIDPService(svc *idp.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.idpSvc = svc
	}
}

func WithSchemasService(svc *schemas.Service) RouterOption {
	return func(c *v0RouterConfig) {
		c.schemaSvc = svc
	}
}

func WithOpenFGA(openfga openfga.OpenFGAClientInterface) RouterOption {
	return func(c *v0RouterConfig) {
		c.openfga = openfga
	}
}

func WithKratos(admin, public *ik.Client) RouterOption {
	return func(c *v0RouterConfig) {
		c.kratosIdentityAPI = admin.IdentityAPI()
		c.kratosFrontendAPI = public.FrontendAPI()
	}
}

func WithO11yConfig(cfg config.O11yConfigInterface) RouterOption {
	return func(c *v0RouterConfig) {
		c.o11yConfig = cfg
	}
}

func NewV0APIRouter(opts ...RouterOption) (http.Handler, error) {
	cfg := &v0RouterConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	o11yConfig := cfg.o11yConfig

	apiRouter := chi.NewRouter()

	identitiesAPI := identities.NewAPI(
		cfg.identitiesSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	clientsAPI := clients.NewAPI(
		cfg.clientsSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	idpAPI := idp.NewAPI(
		cfg.idpSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	schemasAPI := schemas.NewAPI(
		cfg.schemaSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	rolesAPI := roles.NewAPI(
		cfg.rolesSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	groupsAPI := groups.NewAPI(
		cfg.groupsSvc,
		o11yConfig.Tracer(),
		o11yConfig.Monitor(),
		o11yConfig.Logger(),
	)

	if cfg.payloadValidationEnabled {
		validationRegistry := validation.NewRegistry(
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		)
		apiRouter.Use(validationRegistry.ValidationMiddleware)

		identitiesAPI.RegisterValidation(validationRegistry)
		clientsAPI.RegisterValidation(validationRegistry)
		idpAPI.RegisterValidation(validationRegistry)
		schemasAPI.RegisterValidation(validationRegistry)
		rolesAPI.RegisterValidation(validationRegistry)
		groupsAPI.RegisterValidation(validationRegistry)
	}

	gRPCGatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(types.SetHeaderFromMetadataFilter),
		runtime.WithForwardResponseRewriter(types.ForwardErrorResponseRewriter),
	)

	ctx := context.Background()
	err := v0Roles.RegisterRolesServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		roles.NewGrpcHandler(
			cfg.rolesSvc,
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 roles service handler: %w", err)
	}

	err = v0Groups.RegisterGroupsServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		groups.NewGrpcHandler(
			cfg.groupsSvc,
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 groups service handler: %w", err)
	}

	err = v0Status.RegisterStatusServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		status.NewGrpcHandler(
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 status service handler: %w", err)
	}

	err = v0Identities.RegisterIdentitiesServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		identities.NewGrpcHandler(
			cfg.identitiesSvc,
			identities.NewGrpcMapper(o11yConfig.Logger()),
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 identities service handler: %w", err)
	}

	err = v0Idps.RegisterIdpsServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		idp.NewGrpcHandler(
			cfg.idpSvc,
			idp.NewGrpcMapper(o11yConfig.Logger()),
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 idps service handler: %w", err)
	}

	err = v0Schemas.RegisterSchemasServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		schemas.NewGrpcHandler(
			cfg.schemaSvc,
			schemas.NewGrpcMapper(o11yConfig.Logger()),
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 schemas service handler: %w", err)
	}

	err = v0Clients.RegisterClientsServiceHandlerServer(
		ctx,
		gRPCGatewayMux,
		clients.NewGrpcHandler(
			cfg.clientsSvc,
			clients.NewGrpcMapper(o11yConfig.Logger()),
			o11yConfig.Tracer(),
			o11yConfig.Monitor(),
			o11yConfig.Logger(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register v0 clients service handler: %w", err)
	}

	apiRouter.Mount("/identities", gRPCGatewayMux)
	apiRouter.Mount("/roles", gRPCGatewayMux)
	apiRouter.Mount("/groups", gRPCGatewayMux)
	apiRouter.Mount("/idps", gRPCGatewayMux)
	apiRouter.Mount("/schemas", gRPCGatewayMux)
	apiRouter.Mount("/clients", gRPCGatewayMux)
	apiRouter.Mount("/status", gRPCGatewayMux)
	apiRouter.Mount("/version", gRPCGatewayMux)
	apiRouter.Mount("/metrics", metrics.NewAPI(o11yConfig.Logger()).Handler())

	return apiRouter, nil
}
