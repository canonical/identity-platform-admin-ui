// Copyright 2024 Canonical Ltd.

// Package v1 provides HTTP handlers that implement the ReBAC Admin OpenAPI spec.
// This package delegates authorization and data manipulation to user-defined
// "backend"s that implement the designated abstractions.
package v1

import (
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// ReBACAdminBackendParams contains references to user-defined implementation
// of required abstractions, called "backend"s.
type ReBACAdminBackendParams struct {
	GroupsService       GroupsServiceBackend
	GroupsAuthorization GroupsAuthorizationBackend
}

// ReBACAdminBackend represents the ReBAC admin backend as a whole package.
type ReBACAdminBackend struct {
	params  ReBACAdminBackendParams
	service resources.ServerInterface
}

// NewReBACAdminBackend returns a new ReBACAdminBackend instance, configured
// with given backends.
func NewReBACAdminBackend(params ReBACAdminBackendParams) *ReBACAdminBackend {
	return newReBACAdminBackendWithService(params, &service{})
}

// newReBACAdminBackendWithService returns a new ReBACAdminBackend instance, configured
// with given backends and service implementation.
//
// This is intended for internal/test use cases.
func newReBACAdminBackendWithService(params ReBACAdminBackendParams, service resources.ServerInterface) *ReBACAdminBackend {
	return &ReBACAdminBackend{
		params:  params,
		service: service,
	}
}

// Handler returns HTTP handlers implementing the ReBAC Admin OpenAPI spec.
func (b *ReBACAdminBackend) Handler(baseURL string) http.Handler {
	baseURL, _ = strings.CutSuffix(baseURL, "/")
	return resources.HandlerWithOptions(b.service, resources.ChiServerOptions{
		BaseURL:          baseURL + "/v1",
		ErrorHandlerFunc: writeErrorResponse,
	})
}
