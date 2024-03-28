// Copyright 2024 Canonical Ltd.

// Package v1 provides HTTP handlers that implement the ReBAC Admin OpenAPI spec.
// This package delegates authorization and data manipulation to user-defined
// "backend"s that implement the designated abstractions.
package v1

import (
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// ReBACAdminBackendParams contains references to user-defined implementation
// of required abstractions, called "backend"s.
type ReBACAdminBackendParams struct {
	Authenticator            interfaces.Authenticator
	AuthenticatorErrorMapper ErrorResponseMapper

	Identities            interfaces.IdentitiesService
	IdentitiesErrorMapper ErrorResponseMapper

	Roles            interfaces.RolesService
	RolesErrorMapper ErrorResponseMapper

	IdentityProviders            interfaces.IdentityProvidersService
	IdentityProvidersErrorMapper ErrorResponseMapper

	Capabilities            interfaces.CapabilitiesService
	CapabilitiesErrorMapper ErrorResponseMapper

	Entitlements            interfaces.EntitlementsService
	EntitlementsErrorMapper ErrorResponseMapper

	Groups            interfaces.GroupsService
	GroupsErrorMapper ErrorResponseMapper

	Resources            interfaces.ResourcesService
	ResourcesErrorMapper ErrorResponseMapper
}

// ReBACAdminBackend represents the ReBAC admin backend as a whole package.
type ReBACAdminBackend struct {
	params  ReBACAdminBackendParams
	handler resources.ServerInterface
}

// NewReBACAdminBackend returns a new ReBACAdminBackend instance, configured
// with given backends.
func NewReBACAdminBackend(params ReBACAdminBackendParams) *ReBACAdminBackend {
	return newReBACAdminBackendWithService(
		params,
		newHandlerWithValidation(&handler{
			Identities:            params.Identities,
			IdentitiesErrorMapper: params.IdentitiesErrorMapper,

			Roles:            params.Roles,
			RolesErrorMapper: params.RolesErrorMapper,

			IdentityProviders:            params.IdentityProviders,
			IdentityProvidersErrorMapper: params.IdentityProvidersErrorMapper,

			Capabilities:            params.Capabilities,
			CapabilitiesErrorMapper: params.CapabilitiesErrorMapper,

			Entitlements:            params.Entitlements,
			EntitlementsErrorMapper: params.EntitlementsErrorMapper,

			Groups:            params.Groups,
			GroupsErrorMapper: params.GroupsErrorMapper,

			Resources:            params.Resources,
			ResourcesErrorMapper: params.ResourcesErrorMapper,
		}))
}

// newReBACAdminBackendWithService returns a new ReBACAdminBackend instance, configured
// with given backends and service implementation.
//
// This is intended for internal/test use cases.
func newReBACAdminBackendWithService(params ReBACAdminBackendParams, handler resources.ServerInterface) *ReBACAdminBackend {
	return &ReBACAdminBackend{
		params:  params,
		handler: handler,
	}
}

// Handler returns HTTP handlers implementing the ReBAC Admin OpenAPI spec.
func (b *ReBACAdminBackend) Handler(baseURL string) http.Handler {
	baseURL, _ = strings.CutSuffix(baseURL, "/")
	return resources.HandlerWithOptions(b.handler, resources.ChiServerOptions{
		BaseURL: baseURL + "/v1",
		Middlewares: []resources.MiddlewareFunc{
			b.authenticationMiddleware(),
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeErrorResponse(w, err)
		},
	})
}
