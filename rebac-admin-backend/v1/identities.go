// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import "github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"

// IdentitiesServiceBackend defines an abstract backend to handle Identities related operations.
type IdentitiesServiceBackend interface {

	// ListIdentities returns a page of Identity objects of at least `size` elements if available
	ListIdentities(pageToken string, size uint64, filter string) (resources.Identities, error)
	// CreateIdentity creates a single Identity.
	CreateIdentity(identity resources.Identity) (bool, error)

	// GetIdentity returns a single Identity.
	GetIdentity(identityId string) (resources.Identity, error)

	// UpdateIdentity updates an Identity.
	UpdateIdentity(identity resources.Identity) (resources.Identity, error)
	// DeleteIdentity deletes an Identity
	DeleteIdentity(identityId string) (bool, error)

	// GetIdentityGroups returns a page of Groups for identity `identityId`.
	GetIdentityGroups(identityId, pageToken string, size uint64) (resources.Groups, error)
	// PatchIdentityGroups performs addition or removal of a Group to/from an Identity.
	PatchIdentityGroups(identityId string, groupPatches resources.PatchRequestBody) (bool, error)

	// GetIdentityRoles returns a page of Groups for identity `identityId`.
	GetIdentityRoles(identityId, pageToken string, size uint64) (resources.Roles, error)
	// PatchIdentityRoles performs addition or removal of a Role to/from an Identity.
	PatchIdentityRoles(identityId string, rolePatches resources.PatchRequestBody) (bool, error)

	// GetIdentityEntitlements returns a page of Entitlements for identity `identityId`.
	GetIdentityEntitlements(identityId, pageToken string, size uint64) (resources.EntityEntitlements, error)
	// PatchIdentityEntitlements performs addition or removal of an Entitlement to/from an Identity.
	PatchIdentityEntitlements(identityId string, entitlementPatches resources.EntityEntitlementPatchRequestBody) (bool, error)
}

// IdentitiesAuthorizationBackend defines an abstract backend to handle authorization for Identities.
type IdentitiesAuthorizationBackend interface {
}
