// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// PostRoles validates request body for the PostRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PostRoles(w http.ResponseWriter, r *http.Request) {
	body := &resources.Role{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PostRoles(w, r)
	})
}

// PutRolesItem validates request body for the PutRolesItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutRolesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.Role{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		if body.Id == nil || id != *body.Id {
			writeErrorResponse(w, NewRequestBodyValidationError("role ID from path does not match the Role object"))
			return
		}
		v.ServerInterface.PutRolesItem(w, r, id)
	})
}

// PatchRolesItemEntitlements validates request body for the PatchRolesItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchRolesItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.GroupEntitlementsPatchRequestBody{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PatchRolesItemEntitlements(w, r, id)
	})
}
