// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// CapabilitiesService defines an abstract backend to handle capabilities related operations.
type CapabilitiesService interface {
	ListCapabilities(ctx context.Context) (*resources.PaginatedResponse[resources.Capability], error)
}

type CapabilitiesAuthorization interface{}
