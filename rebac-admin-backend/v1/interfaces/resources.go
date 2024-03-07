// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// ResourcesService defines an abstract backend to handle Resources related operations.
type ResourcesService interface {
	// ListResources returns a page of Resource objects of at least `size` elements if available.
	ListResources(ctx context.Context, params *resources.GetResourcesParams) (*resources.Resources, error)
}

// ResourcesAuthorization defines an abstract backend to handle authorization for Resources.
type ResourcesAuthorization interface {
}
