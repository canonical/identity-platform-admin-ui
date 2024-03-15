// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package web

import (
	"context"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	fga "github.com/openfga/go-sdk"
	openfga "github.com/openfga/go-sdk"
)

type OpenFGAClientInterface interface {
	ReadModel(context.Context) (*fga.AuthorizationModel, error)
	CompareModel(context.Context, fga.AuthorizationModel) (bool, error)
	WriteTuple(context.Context, string, string, string) error
	DeleteTuple(context.Context, string, string, string) error
	Check(context.Context, string, string, string) (bool, error)
	ListObjects(context.Context, string, string, string) ([]string, error)
	WriteTuples(context.Context, ...ofga.Tuple) error
	DeleteTuples(context.Context, ...ofga.Tuple) error
	BatchCheck(context.Context, ...ofga.Tuple) (bool, error)
	ReadTuples(context.Context, string, string, string, string) (*openfga.ReadResponse, error)
}
