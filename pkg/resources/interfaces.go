// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package resources

import (
	"context"

	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

type OpenFGAStoreInterface interface {
	ListPermissionsWithFilters(context.Context, string, ...ofga.ListPermissionsFiltersInterface) ([]ofga.Permission, map[string]string, error)
}
