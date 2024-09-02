// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package capabilities_test

import (
	"reflect"
	"testing"

	capabilities "github.com/canonical/identity-platform-admin-ui/pkg/capabilities"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
)

func TestV1ServiceListCapabilities(t *testing.T) {

	tests := []struct {
		name     string
		expected []resources.Capability
	}{
		{
			name: "validate endpoints",
			expected: []resources.Capability{
				{Endpoint: "/api/v1/capabilities", Methods: []resources.CapabilityMethods{"GET"}},
				{Endpoint: "/api/v1/authentication/providers", Methods: []resources.CapabilityMethods{"GET"}},
				{Endpoint: "/api/v1/authentication", Methods: []resources.CapabilityMethods{"GET", "POST"}},
				{Endpoint: "/api/v1/authentication/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
				{Endpoint: "/api/v1/identities", Methods: []resources.CapabilityMethods{"GET", "POST"}},
				{Endpoint: "/api/v1/identities/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
				{Endpoint: "/api/v1/identities/{id}/groups", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/identities/{id}/roles", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/identities/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/groups", Methods: []resources.CapabilityMethods{"GET", "POST"}},
				{Endpoint: "/api/v1/groups/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
				{Endpoint: "/api/v1/groups/{id}/identities", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/groups/{id}/roles", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/groups/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/roles", Methods: []resources.CapabilityMethods{"GET", "POST"}},
				{Endpoint: "/api/v1/roles/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
				{Endpoint: "/api/v1/roles/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
				{Endpoint: "/api/v1/entitlements", Methods: []resources.CapabilityMethods{"GET"}},
				{Endpoint: "/api/v1/entitlements/raw", Methods: []resources.CapabilityMethods{"GET"}},
				{Endpoint: "/api/v1/resources", Methods: []resources.CapabilityMethods{"GET"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := capabilities.NewV1Service()

			caps, err := svc.ListCapabilities()

			if !reflect.DeepEqual(test.expected, caps) || err != nil {
				t.Errorf("expected capabilities to be %v got %v", test.expected, caps)
			}

		})
	}
}
