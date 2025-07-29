// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package v2

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/identity-platform-admin-ui/internal/config"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

type RouterOption func(*v2RouterConfig)

type v2RouterConfig struct {
	contextPath              string
	payloadValidationEnabled bool

	oauth2  *authentication.Config
	openfga openfga.OpenFGAClientInterface

	o11yConfig config.O11yConfigInterface
}

func WithContextPath(path string) RouterOption {
	return func(c *v2RouterConfig) {
		c.contextPath = path
	}
}

func WithOAuth2Config(oauth2Config *authentication.Config) RouterOption {
	return func(c *v2RouterConfig) {
		c.oauth2 = oauth2Config
	}
}
func WithPayloadValidation(enabled bool) RouterOption {
	return func(c *v2RouterConfig) {
		c.payloadValidationEnabled = enabled
	}
}

func WithOpenFGA(openfga openfga.OpenFGAClientInterface) RouterOption {
	return func(c *v2RouterConfig) {
		c.openfga = openfga
	}
}

func WithO11yConfig(cfg config.O11yConfigInterface) RouterOption {
	return func(c *v2RouterConfig) {
		c.o11yConfig = cfg
	}
}

func NewV2APIRouter(opts ...RouterOption) (http.Handler, error) {
	cfg := &v2RouterConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	apiRouter := chi.NewRouter()

	return apiRouter, nil
}
