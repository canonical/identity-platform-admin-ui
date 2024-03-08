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
	Identities              interfaces.IdentitiesService
	IdentitiesAuthorization interfaces.IdentitiesAuthorization
	IdentitiesErrorMapper   ErrorResponseMapper

	Roles              interfaces.RolesService
	RolesAuthorization interfaces.RolesAuthorization
	RolesErrorMapper   ErrorResponseMapper

	IdentityProviders              interfaces.IdentityProvidersService
	IdentityProvidersAuthorization interfaces.IdentityProvidersAuthorization
	IdentityProvidersErrorMapper   ErrorResponseMapper

	Capabilities              interfaces.CapabilitiesService
	CapabilitiesAuthorization interfaces.CapabilitiesAuthorization
	CapabilitiesErrorMapper   ErrorResponseMapper

	Entitlements              interfaces.EntitlementsService
	EntitlementsAuthorization interfaces.EntitlementsAuthorization
	EntitlementsErrorMapper   ErrorResponseMapper

	Groups              interfaces.GroupsService
	GroupsAuthorization interfaces.GroupsAuthorization
	GroupsErrorMapper   ErrorResponseMapper

	Resources              interfaces.ResourcesService
	ResourcesAuthorization interfaces.ResourcesAuthorization
	ResourcesErrorMapper   ErrorResponseMapper
}

// ReBACAdminBackend represents the ReBAC admin backend as a whole package.
type ReBACAdminBackend struct {
	params  ReBACAdminBackendParams
	handler resources.ServerInterface
}

// NewReBACAdminBackend returns a new ReBACAdminBackend instance, configured
// with given backends.
func NewReBACAdminBackend(params ReBACAdminBackendParams) *ReBACAdminBackend {
	return newReBACAdminBackendWithService(params, &handler{
		Identities:              params.Identities,
		IdentitiesAuthorization: params.IdentitiesAuthorization,
		IdentitiesErrorMapper:   params.IdentitiesErrorMapper,

		Roles:              params.Roles,
		RolesAuthorization: params.RolesAuthorization,
		RolesErrorMapper:   params.RolesErrorMapper,

		IdentityProviders:              params.IdentityProviders,
		IdentityProvidersAuthorization: params.IdentityProvidersAuthorization,
		IdentityProvidersErrorMapper:   params.IdentityProvidersErrorMapper,

		Capabilities:              params.Capabilities,
		CapabilitiesAuthorization: params.CapabilitiesAuthorization,
		CapabilitiesErrorMapper:   params.CapabilitiesErrorMapper,

		Entitlements:              params.Entitlements,
		EntitlementsAuthorization: params.EntitlementsAuthorization,
		EntitlementsErrorMapper:   params.EntitlementsErrorMapper,

		Groups:              params.Groups,
		GroupsAuthorization: params.GroupsAuthorization,
		GroupsErrorMapper:   params.GroupsErrorMapper,

		Resources:              params.Resources,
		ResourcesAuthorization: params.ResourcesAuthorization,
		ResourcesErrorMapper:   params.ResourcesErrorMapper,
	})
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
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeErrorResponse(w, err)
		},
	})
}
