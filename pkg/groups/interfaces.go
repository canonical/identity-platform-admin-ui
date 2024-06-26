// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"

	"github.com/openfga/go-sdk/client"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
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
	ListIdentities(context.Context, string, string) ([]string, string, error)
	AssignIdentities(context.Context, string, ...string) error
	RemoveIdentities(context.Context, string, ...string) error
}

// OpenFGAClientInterface is the interface used to decouple the OpenFGA store implementation
type OpenFGAClientInterface interface {
	ListObjects(context.Context, string, string, string) ([]string, error)
	ReadTuples(context.Context, string, string, string, string) (*client.ClientReadResponse, error)
	WriteTuples(context.Context, ...ofga.Tuple) error
	DeleteTuples(context.Context, ...ofga.Tuple) error
	Check(context.Context, string, string, string, ...ofga.Tuple) (bool, error)
}
