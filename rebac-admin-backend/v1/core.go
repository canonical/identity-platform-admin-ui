// Copyright 2024 Canonical Ltd.

// Package v1 provides HTTP handlers that implement the ReBAC Admin OpenAPI spec.
// This package delegates authorization and data manipulation to user-defined
// "backend"s that implement the designated abstractions.
package v1

import (
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// ReBACAdminBackendParams contains references to user-defined implementation
// of required abstractions, called "backend"s.
type ReBACAdminBackendParams struct {
	IdentitiesService       interfaces.IdentitiesService
	IdentitiesAuthorization interfaces.IdentitiesAuthorization
	IdentitiesErrorMapper   ErrorResponseMapper

	RolesService       interfaces.RolesService
	RolesAuthorization interfaces.RolesAuthorization
	RolesErrorMapper   ErrorResponseMapper

	GroupsService       interfaces.GroupsService
	GroupsAuthorization interfaces.GroupsAuthorization
	GroupsErrorMapper   ErrorResponseMapper

	Resources              interfaces.ResourcesService
	ResourcesAuthorization interfaces.ResourcesAuthorization
	ResourcesErrorMapper   ErrorResponseMapper
}

// ReBACAdminBackend represents the ReBAC admin backend as a whole package.
type ReBACAdminBackend struct {
	params  ReBACAdminBackendParams
	handler resources.ServerInterface
}

// NewReBACAdminBackend returns a new ReBACAdminBackend instance, configured
// with given backends.
func NewReBACAdminBackend(params ReBACAdminBackendParams) *ReBACAdminBackend {
	return newReBACAdminBackendWithService(params, &handler{})
}

// newReBACAdminBackendWithService returns a new ReBACAdminBackend instance, configured
// with given backends and service implementation.
//
// This is intended for internal/test use cases.
func newReBACAdminBackendWithService(params ReBACAdminBackendParams, handler resources.ServerInterface) *ReBACAdminBackend {
	return &ReBACAdminBackend{
		params:  params,
		handler: handler,
	}
}

// Handler returns HTTP handlers implementing the ReBAC Admin OpenAPI spec.
func (b *ReBACAdminBackend) Handler(baseURL string) http.Handler {
	baseURL, _ = strings.CutSuffix(baseURL, "/")
	return resources.HandlerWithOptions(b.handler, resources.ChiServerOptions{
		BaseURL: baseURL + "/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeErrorResponse(w, err)
		},
	})
}
