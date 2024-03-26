// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetRoles returns the list of known roles.
// (GET /roles)
func (h handler) GetRoles(w http.ResponseWriter, req *http.Request, params resources.GetRolesParams) {
	ctx := req.Context()

	roles, err := h.Roles.ListRoles(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	response := resources.GetRolesResponse{
		Links:  resources.NewResponseLinks[resources.Role](req.URL, roles),
		Meta:   roles.Meta,
		Data:   roles.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PostRoles adds a new role.
// (POST /roles)
func (h handler) PostRoles(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := getRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	role, ok := body.(*resources.Role)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.Roles.CreateRole(ctx, role)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusCreated, result)
}

// DeleteRolesItem deletes the specified role.
// (DELETE /roles/{id})
func (h handler) DeleteRolesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	_, err := h.Roles.DeleteRole(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetRolesItem returns the role identified by the provided ID.
// (GET /roles/{id})
func (h handler) GetRolesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	role, err := h.Roles.GetRole(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, role)
}

// PutRolesItem updates the role identified by the provided ID.
// (PUT /roles/{id})
func (h handler) PutRolesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := getRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	role, ok := body.(*resources.Role)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.Roles.UpdateRole(ctx, role)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, result)
}

// GetRolesItemEntitlements returns the list of entitlements for a role identified by the provided ID.
// (GET /roles/{id}/entitlements)
func (h handler) GetRolesItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params resources.GetRolesItemEntitlementsParams) {
	ctx := req.Context()

	entitlements, err := h.Roles.GetRoleEntitlements(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	response := resources.GetIdentityEntitlementsResponse{
		Links:  resources.NewResponseLinks[resources.EntityEntitlement](req.URL, entitlements),
		Meta:   entitlements.Meta,
		Data:   entitlements.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchRolesItemEntitlements Adds or removes entitlements to/from a role.
// (PATCH /roles/{id}/entitlements)
func (h handler) PatchRolesItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := getRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	roleEntitlements, ok := body.(*resources.RoleEntitlementsPatchRequestBody)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	_, err = h.Roles.PatchRoleEntitlements(ctx, id, roleEntitlements.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.RolesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
