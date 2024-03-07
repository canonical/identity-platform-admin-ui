// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// IdentityProviderService defines an abstract backend to handle Roles related operations.
type IdentityProviderService interface {

	// ListAvailableIdentityProviders returns the static list of supported identity providers.
	ListAvailableIdentityProviders(ctx context.Context, params *resources.GetAvailableIdentityProvidersParams) (*resources.AvailableIdentityProviders, error)

	// ListIdentityProviders returns a list of registered identity providers configurations.
	ListIdentityProviders(ctx context.Context, params *resources.GetIdentityProvidersParams) (*resources.IdentityProviders, error)

	// RegisterConfiguration register a new authentication provider configuration.
	RegisterConfiguration(ctx context.Context, provider *resources.IdentityProvider) (*resources.IdentityProvider, error)

	// DeleteConfiguration removes an authentication provider configuration identified by `id`.
	DeleteConfiguration(ctx context.Context, id string) (bool, error)

	// GetConfiguration returns the authentication provider configuration identified by `id`.
	GetConfiguration(ctx context.Context, id string) (*resources.IdentityProvider, error)

	// UpdateConfiguration update the authentication provider configuration identified by `id`.
	UpdateConfiguration(ctx context.Context, provider *resources.IdentityProvider) (*resources.IdentityProvider, error)
}

// IdentityProviderAuthorization defines an abstract backend to handle authorization for Roles.
type IdentityProviderAuthorization interface {
}
