// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"

	"github.com/openfga/go-sdk/client"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

// ServiceInterface is the interface that each business logic service needs to implement
type ServiceInterface interface {
	ListRoles(context.Context, string) ([]string, error)
	GetRole(context.Context, string, string) (*Role, error)
	CreateRole(context.Context, string, string) (*Role, error)
	DeleteRole(context.Context, string) error
	ListRoleGroups(context.Context, string) ([]string, error)
	ListPermissions(context.Context, string, map[string]string) ([]string, map[string]string, error)
	AssignPermissions(context.Context, string, ...Permission) error
	RemovePermissions(context.Context, string, ...Permission) error
}

// OpenFGAClientInterface is the interface used to decouple the OpenFGA store implementation
type OpenFGAClientInterface interface {
	ListObjects(context.Context, string, string, string) ([]string, error)
	ListUsers(context.Context, string, string, string) ([]string, error)
	ReadTuples(context.Context, string, string, string, string) (*client.ClientReadResponse, error)
	WriteTuples(context.Context, ...ofga.Tuple) error
	DeleteTuples(context.Context, ...ofga.Tuple) error
	Check(context.Context, string, string, string, ...ofga.Tuple) (bool, error)
}
