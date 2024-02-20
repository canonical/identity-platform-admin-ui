// Copyright 2024 Canonical Ltd.

package v1

import "github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"

type service struct {
	// TODO(CSS-7311): the Unimplemented struct should be removed from here after all
	// endpoints are implemented.
	resources.Unimplemented

	GroupsService       GroupsServiceBackend
	GroupsAuthorization GroupsAuthorizationBackend
}
