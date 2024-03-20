// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetGroups returns the list of known groups.
// (GET /groups)
func (h handler) GetGroups(w http.ResponseWriter, req *http.Request, params resources.GetGroupsParams) {
	ctx := req.Context()

	groups, err := h.Groups.ListGroups(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	response := resources.GetGroupsResponse{
		Links:  resources.NewResponseLinks[resources.Group](req.URL, groups),
		Meta:   groups.Meta,
		Data:   groups.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PostGroups adds a new group.
// (POST /groups)
func (h handler) PostGroups(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	group, err := getRequestBodyFromContext[resources.Group](req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	result, err := h.Groups.CreateGroup(ctx, group)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusCreated, result)
}

// DeleteGroupsItem deletes the specified group identified by the provided ID.
// (DELETE /groups/{id})
func (h handler) DeleteGroupsItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	_, err := h.Groups.DeleteGroup(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetGroupsItem returns the group identified by the provided ID.
// (GET /groups/{id})
func (h handler) GetGroupsItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	group, err := h.Groups.GetGroup(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, group)
}

// PutGroupsItem updates the group identified by the provided ID.
// (PUT /groups/{id})
func (h handler) PutGroupsItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	group, err := getRequestBodyFromContext[resources.Group](req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	result, err := h.Groups.UpdateGroup(ctx, group)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, result)
}

// GetGroupsItemEntitlements returns the list of entitlements for a group identified by the provided ID.
// (GET /groups/{id}/entitlements)
func (h handler) GetGroupsItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params resources.GetGroupsItemEntitlementsParams) {
	ctx := req.Context()

	entitlements, err := h.Groups.GetGroupEntitlements(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	response := resources.GetGroupEntitlementsResponse{
		Links:  resources.NewResponseLinks[resources.EntityEntitlement](req.URL, entitlements),
		Meta:   entitlements.Meta,
		Data:   entitlements.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchGroupsItemEntitlements Adds or removes entitlements to/from a group identified by the provided ID.
// (PATCH /groups/{id}/entitlements)
func (h handler) PatchGroupsItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	groupEntitlements, err := getRequestBodyFromContext[resources.GroupEntitlementsPatchRequestBody](req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	_, err = h.Groups.PatchGroupEntitlements(ctx, id, groupEntitlements.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetGroupsItemIdentities returns the list of identities within a group identified by given id.
// (GET /groups/{id}/identities)
func (h handler) GetGroupsItemIdentities(w http.ResponseWriter, req *http.Request, id string, params resources.GetGroupsItemIdentitiesParams) {
	ctx := req.Context()

	identities, err := h.Groups.GetGroupIdentities(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	response := resources.GetGroupIdentitiesResponse{
		Links:  resources.NewResponseLinks[resources.Identity](req.URL, identities),
		Meta:   identities.Meta,
		Data:   identities.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchGroupsItemIdentities adds or removes identities to/from the group identified by given ID.
// (PATCH /groups/{id}/identities)
func (h handler) PatchGroupsItemIdentities(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	groupIdentities, err := getRequestBodyFromContext[resources.GroupIdentitiesPatchRequestBody](req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	_, err = h.Groups.PatchGroupIdentities(ctx, id, groupIdentities.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetGroupsItemRoles returns the list of roles assigned to a group identified by given ID.
// (GET /groups/{id}/roles)
func (h handler) GetGroupsItemRoles(w http.ResponseWriter, req *http.Request, id string, params resources.GetGroupsItemRolesParams) {
	ctx := req.Context()

	roles, err := h.Groups.GetGroupRoles(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	response := resources.GetIdentityRolesResponse{
		Links:  resources.NewResponseLinks[resources.Role](req.URL, roles),
		Meta:   roles.Meta,
		Data:   roles.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchGroupsItemRoles Add or remove roles assigned to/from a group identified by given ID.
// (PATCH /groups/{id}/roles)
func (h handler) PatchGroupsItemRoles(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	groupRoles, err := getRequestBodyFromContext[resources.GroupRolesPatchRequestBody](req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	_, err = h.Groups.PatchGroupRoles(ctx, id, groupRoles.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.GroupsErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
