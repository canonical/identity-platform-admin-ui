// Copyright 2024 Canonical Ltd.

package v1

import (
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
)

type handler struct {
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
