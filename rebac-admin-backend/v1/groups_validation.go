// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetGroups validates request body for the GetGroups method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroups(w http.ResponseWriter, r *http.Request, params resources.GetGroupsParams) {
	v.handler.GetGroups(w, r, params)
}

// PostGroups validates request body for the PostGroups method and delegates to the underlying handler.
func (v handlerWithValidation) PostGroups(w http.ResponseWriter, r *http.Request) {
	setRequestBodyInContext[resources.Group](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.Group) {
		if err := validateGroup(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.handler.PostGroups(w, r)
	})
}

// DeleteGroupsItem validates request body for the DeleteGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) DeleteGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	v.handler.DeleteGroupsItem(w, r, id)
}

// GetGroupsItem validates request body for the GetGroupsItem method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroupsItem(w http.ResponseWriter, r *http.Request, id string) {
	v.handler.GetGroupsItem(w, r, id)
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
		v.handler.PutGroupsItem(w, r, id)
	})
}

// GetGroupsItemEntitlements validates request body for the GetGroupsItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroupsItemEntitlements(w http.ResponseWriter, r *http.Request, id string, params resources.GetGroupsItemEntitlementsParams) {
	v.handler.GetGroupsItemEntitlements(w, r, id, params)
}

// PatchGroupsItemEntitlements validates request body for the PatchGroupsItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupEntitlementsPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupEntitlementsPatchRequestBody) {
		if err := validateGroupEntitlementsPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.handler.PatchGroupsItemEntitlements(w, r, id)
	})
}

// GetGroupsItemIdentities validates request body for the GetGroupsItemIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroupsItemIdentities(w http.ResponseWriter, r *http.Request, id string, params resources.GetGroupsItemIdentitiesParams) {
	v.handler.GetGroupsItemIdentities(w, r, id, params)
}

// PatchGroupsItemIdentities validates request body for the PatchGroupsItemIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemIdentities(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupIdentitiesPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupIdentitiesPatchRequestBody) {
		if err := validateGroupIdentitiesPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.handler.PatchGroupsItemIdentities(w, r, id)
	})
}

// GetGroupsItemRoles validates request body for the GetGroupsItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) GetGroupsItemRoles(w http.ResponseWriter, r *http.Request, id string, params resources.GetGroupsItemRolesParams) {
	v.handler.GetGroupsItemRoles(w, r, id, params)
}

// PatchGroupsItemRoles validates request body for the PatchGroupsItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PatchGroupsItemRoles(w http.ResponseWriter, r *http.Request, id string) {
	setRequestBodyInContext[resources.GroupRolesPatchRequestBody](w, r, func(w http.ResponseWriter, r *http.Request, body *resources.GroupRolesPatchRequestBody) {
		if err := validateGroupRolesPatchRequestBody(body); err != nil {
			writeErrorResponse(w, err)
			return
		}
		v.handler.PatchGroupsItemRoles(w, r, id)
	})
}
