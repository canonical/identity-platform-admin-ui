// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package capabilities

import "github.com/canonical/rebac-admin-ui-handlers/v1/resources"

type V1Service struct{}

const APIPrefix = "/api/v1"

func (s *V1Service) ListCapabilities() ([]resources.Capability, error) {
	result := []resources.Capability{
		// {Endpoint: APIPrefix + "/swagger.json", Methods: []resources.CapabilityMethods{"GET"}},
		{Endpoint: APIPrefix + "/capabilities", Methods: []resources.CapabilityMethods{"GET"}},
		{Endpoint: APIPrefix + "/authentication/providers", Methods: []resources.CapabilityMethods{"GET"}},
		{Endpoint: APIPrefix + "/authentication", Methods: []resources.CapabilityMethods{"GET", "POST"}},
		{Endpoint: APIPrefix + "/authentication/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
		{Endpoint: APIPrefix + "/identities", Methods: []resources.CapabilityMethods{"GET", "POST"}},
		{Endpoint: APIPrefix + "/identities/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
		{Endpoint: APIPrefix + "/identities/{id}/groups", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/identities/{id}/roles", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/identities/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/groups", Methods: []resources.CapabilityMethods{"GET", "POST"}},
		{Endpoint: APIPrefix + "/groups/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
		{Endpoint: APIPrefix + "/groups/{id}/identities", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/groups/{id}/roles", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/groups/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/roles", Methods: []resources.CapabilityMethods{"GET", "POST"}},
		{Endpoint: APIPrefix + "/roles/{id}", Methods: []resources.CapabilityMethods{"GET", "PUT", "DELETE"}},
		{Endpoint: APIPrefix + "/roles/{id}/entitlements", Methods: []resources.CapabilityMethods{"GET", "PATCH"}},
		{Endpoint: APIPrefix + "/entitlements", Methods: []resources.CapabilityMethods{"GET"}},
		{Endpoint: APIPrefix + "/entitlements/raw", Methods: []resources.CapabilityMethods{"GET"}},
		{Endpoint: APIPrefix + "/resources", Methods: []resources.CapabilityMethods{"GET"}},
	}

	return result, nil
}

func NewV1Service() *V1Service {
	s := new(V1Service)

	return s
}
