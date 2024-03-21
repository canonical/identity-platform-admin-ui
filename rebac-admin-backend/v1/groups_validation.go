// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// PostGroups validates request body for the PostGroups method and delegates to the underlying handler.
func (v handlerWithValidation) PostGroups(w http.ResponseWriter, r *http.Request) {
	setRequestBodyInContext[resources.Group](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.Group) {
		if err := validateGroup(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.ServerInterface.PostGroups(w, r)
	})
}

// DeleteGroupsItem validates request body for the DeleteGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) DeleteGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	v.ServerInterface.DeleteGroupsItem(w, r, id)
}

// GetGroupsItem validates request body for the GetGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	v.ServerInterface.GetGroupsItem(w, r, id)
}

// PutGroupsItem validates request body for the PutGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.Group](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.Group) {
		if err := validateGroup(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		if body.Id == nil || id != *body.Id {
			writeErrorResponse(w, NewRequestBodyValidationError("group ID from path does not match the Group object"))
			return
		}
		v.ServerInterface.PutGroupsItem(w, r, id)
	})
}

// PatchGroupsItemEntitlements validates request body for the PatchGroupsItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupEntitlementsPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupEntitlementsPatchRequestBody) {
		if err := validateGroupEntitlementsPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.ServerInterface.PatchGroupsItemEntitlements(w, r, id)
	})
}

// PatchGroupsItemIdentities validates request body for the PatchGroupsItemIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemIdentities(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupIdentitiesPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupIdentitiesPatchRequestBody) {
		if err := validateGroupIdentitiesPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.ServerInterface.PatchGroupsItemIdentities(w, r, id)
	})
}

// PatchGroupsItemRoles validates request body for the PatchGroupsItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemRoles(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupRolesPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupRolesPatchRequestBody) {
		if err := validateGroupRolesPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.ServerInterface.PatchGroupsItemRoles(w, r, id)
	})
}
