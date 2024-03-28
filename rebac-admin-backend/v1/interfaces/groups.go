// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GroupsService defines an abstract backend to handle Groups related operations.
type GroupsService interface {

	// ListGroups returns a page of Group objects of at least `size` elements if available.
	ListGroups(ctx context.Context, params *resources.GetGroupsParams) (*resources.PaginatedResponse[resources.Group], error)
	// CreateGroup creates a single Group.
	CreateGroup(ctx context.Context, group *resources.Group) (*resources.Group, error)

	// GetGroup returns a single Group identified by `groupId`.
	GetGroup(ctx context.Context, groupId string) (*resources.Group, error)

	// UpdateGroup updates a Group.
	UpdateGroup(ctx context.Context, group *resources.Group) (*resources.Group, error)
	// DeleteGroup deletes a Group identified by `groupId`.
	// returns (true, nil) in case the group was successfully deleted.
	// returns (false, error) in case something went wrong.
	// implementors may want to return (false, nil) for idempotency cases.
	DeleteGroup(ctx context.Context, groupId string) (bool, error)

	// GetGroupIdentities returns a page of identities in a Group identified by `groupId`.
	GetGroupIdentities(ctx context.Context, groupId string, params *resources.GetGroupsItemIdentitiesParams) (*resources.PaginatedResponse[resources.Identity], error)
	// PatchGroupIdentities performs addition or removal of identities to/from a Group identified by `groupId`.
	PatchGroupIdentities(ctx context.Context, groupId string, identityPatches []resources.GroupIdentitiesPatchItem) (bool, error)

	// GetGroupRoles returns a page of Roles for Group `groupId`.
	GetGroupRoles(ctx context.Context, groupId string, params *resources.GetGroupsItemRolesParams) (*resources.PaginatedResponse[resources.Role], error)
	// PatchGroupRoles performs addition or removal of a Role to/from a Group identified by `groupId`.
	PatchGroupRoles(ctx context.Context, groupId string, rolePatches []resources.GroupRolesPatchItem) (bool, error)

	// GetGroupEntitlements returns a page of Entitlements for Group `groupId`.
	GetGroupEntitlements(ctx context.Context, groupId string, params *resources.GetGroupsItemEntitlementsParams) (*resources.PaginatedResponse[resources.EntityEntitlement], error)
	// PatchGroupEntitlements performs addition or removal of an Entitlement to/from a Group identified by `groupId`.
	PatchGroupEntitlements(ctx context.Context, groupId string, entitlementPatches []resources.GroupEntitlementsPatchItem) (bool, error)
}
