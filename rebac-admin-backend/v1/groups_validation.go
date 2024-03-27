// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// PostGroups validates request body for the PostGroups method and delegates to the underlying handler.
func (v handlerWithValidation) PostGroups(w http.ResponseWriter, r *http.Request) {
	body := &resources.Group{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PostGroups(w, r)
	})
}

// PutGroupsItem validates request body for the PutGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.Group{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		if body.Id == nil || id != *body.Id {
			writeErrorResponse(w, NewRequestBodyValidationError("group ID from path does not match the Group object"))
			return
		}
		v.ServerInterface.PutGroupsItem(w, r, id)
	})
}

// PatchGroupsItemEntitlements validates request body for the PatchGroupsItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.GroupEntitlementsPatchRequestBody{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PatchGroupsItemEntitlements(w, r, id)
	})
}

// PatchGroupsItemIdentities validates request body for the PatchGroupsItemIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemIdentities(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.GroupIdentitiesPatchRequestBody{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PatchGroupsItemIdentities(w, r, id)
	})
}

// PatchGroupsItemRoles validates request body for the PatchGroupsItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemRoles(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.GroupRolesPatchRequestBody{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PatchGroupsItemRoles(w, r, id)
	})
}
