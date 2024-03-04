// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// IdentitiesService defines an abstract backend to handle Identities related operations.
type IdentitiesService interface {

	// ListIdentities returns a page of Identity objects of at least `size` elements if available
	ListIdentities(ctx context.Context, params *resources.GetIdentitiesParams) (*resources.Identities, error)
	// CreateIdentity creates a single Identity.
	CreateIdentity(ctx context.Context, identity *resources.Identity) (*resources.Identity, error)

	// GetIdentity returns a single Identity.
	GetIdentity(ctx context.Context, identityId string) (*resources.Identity, error)

	// UpdateIdentity updates an Identity.
	UpdateIdentity(ctx context.Context, identity *resources.Identity) (*resources.Identity, error)
	// DeleteIdentity deletes an Identity
	// returns (true, nil) in case an identity was successfully delete
	// return (false, error) in case something went wrong
	// implementors may want to return (false, nil) for idempotency cases
	DeleteIdentity(ctx context.Context, identityId string) (bool, error)

	// GetIdentityGroups returns a page of Groups for identity `identityId`.
	GetIdentityGroups(ctx context.Context, identityId string, params *resources.GetIdentitiesItemGroupsParams) (*resources.Groups, error)
	// PatchIdentityGroups performs addition or removal of a Group to/from an Identity.
	PatchIdentityGroups(ctx context.Context, identityId string, groupPatches []resources.IdentityGroupsPatchItem) (bool, error)

	// GetIdentityRoles returns a page of Groups for identity `identityId`.
	GetIdentityRoles(ctx context.Context, identityId string, params *resources.GetIdentitiesItemRolesParams) (*resources.Roles, error)
	// PatchIdentityRoles performs addition or removal of a Role to/from an Identity.
	PatchIdentityRoles(ctx context.Context, identityId string, rolePatches []resources.IdentityRolesPatchItem) (bool, error)

	// GetIdentityEntitlements returns a page of Entitlements for identity `identityId`.
	GetIdentityEntitlements(ctx context.Context, identityId string, params *resources.GetIdentitiesItemEntitlementsParams) ([]resources.EntityEntitlement, error)
	// PatchIdentityEntitlements performs addition or removal of an Entitlement to/from an Identity.
	PatchIdentityEntitlements(ctx context.Context, identityId string, entitlementPatches []resources.IdentityEntitlementsPatchItem) (bool, error)
}

// IdentitiesAuthorization defines an abstract backend to handle authorization for Identities.
type IdentitiesAuthorization interface {
}
