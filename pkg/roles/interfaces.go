// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
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

// RoleRepositoryInterface implements a data access object with the repository pattern
type RoleRepositoryInterface interface {
	FindRoleByName(context.Context, string) (*Role, error)
	FindRoleByIdAndOwner(context.Context, string, string) (*Role, error)
	FindRoleByNameAndOwner(context.Context, string, string) (*Role, error)
	ListRoles(context.Context, string, int64, int64) ([]string, error)
	ListRoleGroups(context.Context, string, int64, int64) ([]string, error)
	CreateRole(context.Context, string, string) (*Role, error)
	CreateRoleTx(context.Context, string, string) (*Role, storage.TxInterface, error)
	DeleteRoleByName(context.Context, string) (string, error)
	DeleteRoleTx(context.Context, string) (string, storage.TxInterface, error)
}
