// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	r "github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// IdentitiesService defines an abstract backend to handle Identities related operations.
type IdentitiesService interface {

	// ListIdentities returns a page of Identity objects of at least `size` elements if available
	ListIdentities(ctx context.Context, params *r.GetIdentitiesParams) (*r.Identities, error)
	// CreateIdentity creates a single Identity.
	CreateIdentity(ctx context.Context, identity *r.Identity) (*r.Identity, error)

	// GetIdentity returns a single Identity.
	GetIdentity(ctx context.Context, identityId string) (*r.Identity, error)

	// UpdateIdentity updates an Identity.
	UpdateIdentity(ctx context.Context, identity *r.Identity) (*r.Identity, error)
	// DeleteIdentity deletes an Identity
	// returns (true, nil) in case an identity was successfully delete
	// return (false, error) in case something went wrong
	// implementors may want to return (false, nil) for idempotency cases
	DeleteIdentity(ctx context.Context, identityId string) (bool, error)

	// GetIdentityGroups returns a page of Groups for identity `identityId`.
	GetIdentityGroups(ctx context.Context, identityId string, params *r.GetIdentitiesItemGroupsParams) (*r.Groups, error)
	// PatchIdentityGroups performs addition or removal of a Group to/from an Identity.
	PatchIdentityGroups(ctx context.Context, identityId string, groupPatches []r.IdentityGroupsPatchItem) (bool, error)

	// GetIdentityRoles returns a page of Groups for identity `identityId`.
	GetIdentityRoles(ctx context.Context, identityId string, params *r.GetIdentitiesItemRolesParams) (*r.Roles, error)
	// PatchIdentityRoles performs addition or removal of a Role to/from an Identity.
	PatchIdentityRoles(ctx context.Context, identityId string, rolePatches []r.IdentityRolesPatchItem) (bool, error)

	// GetIdentityEntitlements returns a page of Entitlements for identity `identityId`.
	GetIdentityEntitlements(ctx context.Context, identityId string, params *r.GetIdentitiesItemEntitlementsParams) ([]r.EntityEntitlement, error)
	// PatchIdentityEntitlements performs addition or removal of an Entitlement to/from an Identity.
	PatchIdentityEntitlements(ctx context.Context, identityId string, entitlementPatches []r.IdentityEntitlementsPatchItem) (bool, error)
}

// IdentitiesAuthorization defines an abstract backend to handle authorization for Identities.
type IdentitiesAuthorization interface {
}
