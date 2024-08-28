// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"

	kClient "github.com/ory/kratos-client-go"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

type AuthorizerInterface interface {
	SetCreateIdentityEntitlements(context.Context, string) error
	SetDeleteIdentityEntitlements(context.Context, string) error
}

type ServiceInterface interface {
	ListIdentities(context.Context, int64, string, string) (*IdentityData, error)
	GetIdentity(context.Context, string) (*IdentityData, error)
	CreateIdentity(context.Context, *kClient.CreateIdentityBody) (*IdentityData, error)
	UpdateIdentity(context.Context, string, *kClient.UpdateIdentityBody) (*IdentityData, error)
	DeleteIdentity(context.Context, string) (*IdentityData, error)
}

type OpenFGAStoreInterface interface {
	ListAssignedRoles(context.Context, string) ([]string, error)
	ListAssignedGroups(context.Context, string) ([]string, error)
	AssignRoles(context.Context, string, ...string) error
	UnassignRoles(context.Context, string, ...string) error
	AssignGroups(context.Context, string, ...string) error
	UnassignGroups(context.Context, string, ...string) error
	ListPermissions(context.Context, string, map[string]string) ([]ofga.Permission, map[string]string, error)
	AssignPermissions(context.Context, string, ...ofga.Permission) error
	UnassignPermissions(context.Context, string, ...ofga.Permission) error
}
