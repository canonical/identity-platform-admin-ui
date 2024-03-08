// Copyright 2024 Canonical Ltd.

package v1

import (
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
)

type handler struct {
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
