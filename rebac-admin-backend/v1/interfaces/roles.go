// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// RolesService defines an abstract backend to handle Roles related operations.
type RolesService interface {
	// ListRoles returns a page of Role objects of at least `size` elements if available
	ListRoles(ctx context.Context, params *resources.GetRolesParams) (*resources.Roles, error)
	// CreateRole creates a single Role.
	CreateRole(ctx context.Context, Role *resources.Role) (*resources.Role, error)

	// GetRole returns a single Role.
	GetRole(ctx context.Context, RoleId string) (*resources.Role, error)
	// UpdateRole updates an Role.
	UpdateRole(ctx context.Context, Role *resources.Role) (*resources.Role, error)
	// DeleteRole deletes a Role
	// returns (true, nil) in case a Role was successfully deleted
	// return (false, error) in case something went wrong
	// implementors may want to return (false, nil) for idempotency cases
	DeleteRole(ctx context.Context, RoleId string) (bool, error)

	// GetRoleEntitlements returns a page of Entitlements for Role `RoleId`.
	GetRoleEntitlements(ctx context.Context, RoleId string, params *resources.GetRolesItemEntitlementsParams) ([]resources.EntityEntitlement, error)
	// PatchRoleEntitlements performs addition or removal of an Entitlement to/from a Role.
	PatchRoleEntitlements(ctx context.Context, RoleId string, entitlementPatches []resources.RoleEntitlementsPatchItem) (bool, error)
}

// RolesAuthorization defines an abstract backend to handle authorization for Roles.
type RolesAuthorization interface {
}
