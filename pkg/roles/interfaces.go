// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package roles

import (
	"context"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/openfga/go-sdk/client"
)

type ServiceInterface interface {
	ListRoles(context.Context, string) ([]string, error) // list of roles, continuation token, error
	GetRole(context.Context, string, string) (string, error)
	CreateRole(context.Context, string, string) error
	DeleteRole(context.Context, string) error
	ListRoleGroups(context.Context, string, string) ([]string, string, error)
	ListPermissions(context.Context, string, map[string]string) ([]string, map[string]string, error) // list of permissions, {"type":"continuation tokens"}
	AssignPermissions(context.Context, string, ...Permission) error                                  // passing in role id and list of groups
	RemovePermissions(context.Context, string, ...Permission) error
}

type OpenFGAClientInterface interface {
	ListObjects(context.Context, string, string, string) ([]string, error)
	ReadTuples(context.Context, string, string, string, string) (*client.ClientReadResponse, error)
	WriteTuple(context.Context, string, string, string) error
	DeleteTuple(context.Context, string, string, string) error
	WriteTuples(context.Context, ...ofga.Tuple) error
	DeleteTuples(context.Context, ...ofga.Tuple) error
	Check(context.Context, string, string, string) (bool, error)
}
