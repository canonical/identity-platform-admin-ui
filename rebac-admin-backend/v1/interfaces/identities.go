// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import r "github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"

// IdentitiesService defines an abstract backend to handle Identities related operations.
type IdentitiesService interface {

	// ListIdentities returns a page of Identity objects of at least `size` elements if available
	ListIdentities(params *r.GetIdentitiesParams) (*r.Identities, error)
	// CreateIdentity creates a single Identity.
	CreateIdentity(identity *r.Identity) (*r.Identity, error)

	// GetIdentity returns a single Identity.
	GetIdentity(identityId string) (*r.Identity, error)

	// UpdateIdentity updates an Identity.
	UpdateIdentity(identity *r.Identity) (*r.Identity, error)
	// DeleteIdentity deletes an Identity
	// returns (true, nil) in case an identity was successfully delete
	// return (false, error) in case something went wrong
	// implementors may want to return (false, nil) for idempotency cases
	DeleteIdentity(identityId string) (bool, error)

	// GetIdentityGroups returns a page of Groups for identity `identityId`.
	GetIdentityGroups(identityId string, params *r.GetIdentitiesItemGroupsParams) (*r.Groups, error)
	// PatchIdentityGroups performs addition or removal of a Group to/from an Identity.
	PatchIdentityGroups(identityId string, groupPatches *r.PatchRequestBody) (bool, error)

	// GetIdentityRoles returns a page of Groups for identity `identityId`.
	GetIdentityRoles(identityId string, params *r.GetIdentitiesItemRolesParams) (*r.Roles, error)
	// PatchIdentityRoles performs addition or removal of a Role to/from an Identity.
	PatchIdentityRoles(identityId string, rolePatches *r.PatchRequestBody) (bool, error)

	// GetIdentityEntitlements returns a page of Entitlements for identity `identityId`.
	GetIdentityEntitlements(identityId string, params *r.GetIdentitiesItemEntitlementsParams) (*r.EntityEntitlements, error)
	// PatchIdentityEntitlements performs addition or removal of an Entitlement to/from an Identity.
	PatchIdentityEntitlements(identityId string, entitlementPatches *r.EntityEntitlementPatchRequestBody) (bool, error)
}

// IdentitiesAuthorization defines an abstract backend to handle authorization for Identities.
type IdentitiesAuthorization interface {
}
