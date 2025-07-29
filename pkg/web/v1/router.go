// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package v1

import (
	"net/http"

	v1handlers "github.com/canonical/rebac-admin-ui-handlers/v1"

	"github.com/canonical/identity-platform-admin-ui/internal/config"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/pkg/entitlements"
	"github.com/canonical/identity-platform-admin-ui/pkg/groups"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/resources"
	"github.com/canonical/identity-platform-admin-ui/pkg/roles"
)

type RouterOption func(*v1RouterConfig)

type v1RouterConfig struct {
	idpConfig     *idp.Config
	store         *openfga.OpenFGAStore
	rolesSvc      *roles.Service
	groupsSvc     *groups.Service
	identitiesSvc *identities.Service
	idpSvc        *idp.Service
	openfga       openfga.OpenFGAClientInterface

	o11yConfig config.O11yConfigInterface
}

func WithIDPConfig(cfg *idp.Config) RouterOption {
	return func(c *v1RouterConfig) {
		c.idpConfig = cfg
	}
}

func WithStore(store *openfga.OpenFGAStore) RouterOption {
	return func(c *v1RouterConfig) {
		c.store = store
	}
}

func WithRolesService(svc *roles.Service) RouterOption {
	return func(c *v1RouterConfig) {
		c.rolesSvc = svc
	}
}

func WithGroupsService(svc *groups.Service) RouterOption {
	return func(c *v1RouterConfig) {
		c.groupsSvc = svc
	}
}

func WithIdentitiesService(svc *identities.Service) RouterOption {
	return func(c *v1RouterConfig) {
		c.identitiesSvc = svc
	}
}

func WithIDPService(svc *idp.Service) RouterOption {
	return func(c *v1RouterConfig) {
		c.idpSvc = svc
	}
}

func WithOpenFGA(openfga openfga.OpenFGAClientInterface) RouterOption {
	return func(c *v1RouterConfig) {
		c.openfga = openfga
	}
}

func WithO11yConfig(cfg config.O11yConfigInterface) RouterOption {
	return func(c *v1RouterConfig) {
		c.o11yConfig = cfg
	}
}

func NewV1APIRouter(opts ...RouterOption) (http.Handler, error) {
	cfg := &v1RouterConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	o11yConfig := cfg.o11yConfig

	backend, err := v1handlers.NewReBACAdminBackend(
		v1handlers.ReBACAdminBackendParams{
			Resources: resources.NewV1Service(cfg.store, o11yConfig.Tracer(), o11yConfig.Monitor(), o11yConfig.Logger()),
			Roles:     roles.NewV1Service(cfg.rolesSvc),
			Groups:    groups.NewV1Service(cfg.groupsSvc, o11yConfig.Tracer(), o11yConfig.Monitor(), o11yConfig.Logger()),
			Identities: identities.NewV1Service(
				&identities.Config{
					Name:         cfg.idpConfig.Name,
					Namespace:    cfg.idpConfig.Namespace,
					K8s:          cfg.idpConfig.K8s,
					OpenFGAStore: cfg.store,
				},
				cfg.identitiesSvc,
			),
			Entitlements:      entitlements.NewV1Service(cfg.openfga, o11yConfig.Tracer(), o11yConfig.Monitor(), o11yConfig.Logger()),
			IdentityProviders: idp.NewV1Service(cfg.idpSvc),
		},
	)

	if err != nil {
		return nil, err
	}

	return backend.Handler(""), nil
}
