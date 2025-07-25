// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
)

// ServiceInterface is the interface that each business logic service needs to implement
type ServiceInterface interface {
	ListGroups(context.Context, string) ([]string, error) // list of groups, continuation token, error
	GetGroup(context.Context, string, string) (*Group, error)
	CreateGroup(context.Context, string, string) (*Group, error)
	DeleteGroup(context.Context, string) error
	ListRoles(context.Context, string) ([]string, error)
	AssignRoles(context.Context, string, ...string) error
	RemoveRoles(context.Context, string, ...string) error
	ListPermissions(context.Context, string, map[string]string) ([]string, map[string]string, error)
	AssignPermissions(context.Context, string, ...Permission) error
	RemovePermissions(context.Context, string, ...Permission) error
	ListIdentities(context.Context, string) ([]string, error)
	AssignIdentities(context.Context, string, ...string) error
	RemoveIdentities(context.Context, string, ...string) error
	CanAssignRoles(context.Context, string, ...string) (bool, error)
	CanAssignIdentities(context.Context, string, ...string) (bool, error)
}

// GroupRepositoryInterface implements a data access object with the repository pattern
type GroupRepositoryInterface interface {
	FindGroupByName(context.Context, string) (*Group, error)
	FindGroupByIdAndOwner(context.Context, string, string) (*Group, error)
	FindGroupByNameAndOwner(context.Context, string, string) (*Group, error)
	ListGroups(context.Context, string, int64, int64) ([]string, error)
	CreateGroup(context.Context, string, string) (*Group, error)
	CreateGroupTx(context.Context, string, string) (*Group, storage.TxInterface, error)
	DeleteGroupByName(context.Context, string) (string, error)
	DeleteGroupTx(context.Context, string) (string, storage.TxInterface, error)
}
