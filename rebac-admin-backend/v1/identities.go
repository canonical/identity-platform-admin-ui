// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetIdentities returns the list of known identities.
// (GET /identities)
func (h handler) GetIdentities(w http.ResponseWriter, req *http.Request, params resources.GetIdentitiesParams) {
	ctx := req.Context()

	identities, err := h.Identities.ListIdentities(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	response := resources.GetIdentitiesResponse{
		Links:  resources.NewResponseLinks[resources.Identity](req.URL, identities),
		Meta:   identities.Meta,
		Data:   identities.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PostIdentities adds a new identity.
// (POST /identities)
func (h handler) PostIdentities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := GetRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identity, ok := body.(*resources.Identity)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.Identities.CreateIdentity(ctx, identity)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusCreated, result)
}

// DeleteIdentitiesItem deletes the specified identity.
// (DELETE /identities/{id})
func (h handler) DeleteIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	_, err := h.Identities.DeleteIdentity(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetIdentitiesItem returns the identity identified by the provided ID.
// (GET /identities/{id})
func (h handler) GetIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	identity, err := h.Identities.GetIdentity(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, identity)
}

// PutIdentitiesItem updates the identity identified by the provided ID.
// (PUT /identities/{id})
func (h handler) PutIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := GetRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identity, ok := body.(*resources.Identity)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.Identities.UpdateIdentity(ctx, identity)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, result)
}

// GetIdentitiesItemEntitlements returns the list of entitlements for an identity identified by the provided ID.
// (GET /identities/{id}/entitlements)
func (h handler) GetIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemEntitlementsParams) {
	ctx := req.Context()

	entitlements, err := h.Identities.GetIdentityEntitlements(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	response := resources.GetIdentityEntitlementsResponse{
		Links:  resources.NewResponseLinks[resources.EntityEntitlement](req.URL, entitlements),
		Meta:   entitlements.Meta,
		Data:   entitlements.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemEntitlements Adds or removes entitlements to/from an identity.
// (PATCH /identities/{id}/entitlements)
func (h handler) PatchIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := GetRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identityEntitlements, ok := body.(*resources.IdentityEntitlementsPatchRequestBody)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	_, err = h.Identities.PatchIdentityEntitlements(ctx, id, identityEntitlements.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetIdentitiesItemGroups returns the list of groups the identity is a member of.
// (GET /identities/{id}/groups)
func (h handler) GetIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemGroupsParams) {
	ctx := req.Context()

	groups, err := h.Identities.GetIdentityGroups(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	response := resources.GetIdentityGroupsResponse{
		Links:  resources.NewResponseLinks[resources.Group](req.URL, groups),
		Meta:   groups.Meta,
		Data:   groups.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemGroups adds or removes the identity to/from groups.
// (PATCH /identities/{id}/groups)
func (h handler) PatchIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := GetRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identityGroups, ok := body.(*resources.IdentityGroupsPatchRequestBody)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	_, err = h.Identities.PatchIdentityGroups(ctx, id, identityGroups.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetIdentitiesItemRoles returns the list of roles assigned to the identity.
// (GET /identities/{id}/roles)
func (h handler) GetIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemRolesParams) {
	ctx := req.Context()

	roles, err := h.Identities.GetIdentityRoles(ctx, id, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	response := resources.GetIdentityRolesResponse{
		Links:  resources.NewResponseLinks[resources.Role](req.URL, roles),
		Meta:   roles.Meta,
		Data:   roles.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemRoles Add or remove the identity to/from roles.
// (PATCH /identities/{id}/roles)
func (h handler) PatchIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := GetRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identityRoles, ok := body.(*resources.IdentityRolesPatchRequestBody)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	_, err = h.Identities.PatchIdentityRoles(ctx, id, identityRoles.Patches)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentitiesErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
