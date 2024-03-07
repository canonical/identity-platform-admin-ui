// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetIdentities returns the list of known identities.
// (GET /identities)
func (h handler) GetIdentities(w http.ResponseWriter, req *http.Request, params resources.GetIdentitiesParams) {
	ctx := req.Context()

	identities, err := h.Identities.ListIdentities(ctx, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentitiesResponse{
		Data:   identities.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PostIdentities adds a new identity.
// (POST /identities)
func (h handler) PostIdentities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	identity := new(resources.Identity)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(identity); err != nil {
		writeErrorResponse(w, NewValidationError("request doesn't match the expected schema"))
		return
	}

	identity, err := h.Identities.CreateIdentity(ctx, identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, http.StatusCreated, identity)
}

// DeleteIdentitiesItem deletes the specified identity.
// (DELETE /identities/{id})
func (h handler) DeleteIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	_, err := h.Identities.DeleteIdentity(ctx, id)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
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
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, http.StatusOK, identity)
}

// PutIdentitiesItem updates the identity identified by the provided ID.
// (PUT /identities/{id})
func (h handler) PutIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	identity := new(resources.Identity)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(identity); err != nil {
		writeErrorResponse(w, NewValidationError("request doesn't match the expected schema"))
		return
	}

	if id != *identity.Id {
		writeErrorResponse(w, NewValidationError("identity ID from path does not match the Identity object"))
		return
	}

	identity, err := h.Identities.UpdateIdentity(ctx, identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, http.StatusOK, identity)
}

// GetIdentitiesItemEntitlements returns the list of entitlements for an identity identified by the provided ID.
// (GET /identities/{id}/entitlements)
func (h handler) GetIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemEntitlementsParams) {
	ctx := req.Context()

	entitlements, err := h.Identities.GetIdentityEntitlements(ctx, id, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityEntitlementsResponse{
		Data:   entitlements,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemEntitlements Adds or removes entitlements to/from an identity.
// (PATCH /identities/{id}/entitlements)
func (h handler) PatchIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	patchesRequest := new(resources.IdentityEntitlementsPatchRequestBody)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(patchesRequest); err != nil {
		writeErrorResponse(w, NewValidationError("request doesn't match the expected schema"))
		return
	}

	_, err := h.Identities.PatchIdentityEntitlements(ctx, id, patchesRequest.Patches)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
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
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityGroupsResponse{
		Data:   groups.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemGroups adds or removes the identity to/from groups.
// (PATCH /identities/{id}/groups)
func (h handler) PatchIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	patchesRequest := new(resources.IdentityGroupsPatchRequestBody)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(patchesRequest); err != nil {
		writeErrorResponse(w, NewValidationError("request doesn't match the expected schema"))
		return
	}

	_, err := h.Identities.PatchIdentityGroups(ctx, id, patchesRequest.Patches)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
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
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityRolesResponse{
		Data:   roles.Data,
		Status: 200,
	}

	writeResponse(w, http.StatusOK, response)
}

// PatchIdentitiesItemRoles Add or remove the identity to/from roles.
// (PATCH /identities/{id}/roles)
func (h handler) PatchIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	patchesRequest := new(resources.IdentityRolesPatchRequestBody)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(patchesRequest); err != nil {
		writeErrorResponse(w, NewValidationError("request doesn't match the expected schema"))
		return
	}

	_, err := h.Identities.PatchIdentityRoles(ctx, id, patchesRequest.Patches)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	w.WriteHeader(http.StatusOK)
}
