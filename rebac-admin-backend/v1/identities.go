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
	identities, err := h.Identities.ListIdentities(&params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentitiesResponse{
		Data:   identities.Data,
		Status: 200,
	}

	writeResponse(w, 200, response)
}

// PostIdentities adds a new identity.
// (POST /identities)
func (h handler) PostIdentities(w http.ResponseWriter, req *http.Request) {
	identity := new(resources.Identity)
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(identity)
	if err != nil {
		writeErrorResponse(w, NewValidationError("Request doesn't match the expected schema"))
		return
	}

	identity, err = h.Identities.CreateIdentity(identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, 201, identity)
}

// DeleteIdentitiesItem deletes the specified identity.
// (DELETE /identities/{id})
func (h handler) DeleteIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	_, err := h.Identities.DeleteIdentity(id)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	w.WriteHeader(200)
}

// GetIdentitiesItem returns the identity identified by the provided ID.
// (GET /identities/{id})
func (h handler) GetIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	identity, err := h.Identities.GetIdentity(id)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, 200, identity)
}

// PutIdentitiesItem updates the identity identified by the provided ID.
// (PUT /identities/{id})
func (h handler) PutIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	identity := new(resources.Identity)
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(identity)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	if id != *identity.Id {
		writeErrorResponse(w, NewValidationError("Identity ID from path does not match the Identity object"))
		return
	}

	identity, err = h.Identities.UpdateIdentity(identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	writeResponse(w, 200, identity)
}

// GetIdentitiesItemEntitlements returns the list of entitlements for an identity identified by the provided ID.
// (GET /identities/{id}/entitlements)
func (h handler) GetIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemEntitlementsParams) {
	entitlements, err := h.Identities.GetIdentityEntitlements(id, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityEntitlementsResponse{
		Data:   entitlements.Data,
		Status: 200,
	}

	writeResponse(w, 200, response)
}

// PatchIdentitiesItemEntitlements Adds or removes entitlements to/from an identity.
// (PATCH /identities/{id}/entitlements)
func (h handler) PatchIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetIdentitiesItemGroups returns the list of groups the identity is a member of.
// (GET /identities/{id}/groups)
func (h handler) GetIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemGroupsParams) {
	groups, err := h.Identities.GetIdentityGroups(id, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityGroupsResponse{
		Data:   groups.Data,
		Status: 200,
	}

	writeResponse(w, 200, response)
}

// PatchIdentitiesItemGroups adds or removes the identity to/from a group.
// (PATCH /identities/{id}/groups)
func (h handler) PatchIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetIdentitiesItemRoles returns the list of roles assigned to the identity.
// (GET /identities/{id}/roles)
func (h handler) GetIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string, params resources.GetIdentitiesItemRolesParams) {
	roles, err := h.Identities.GetIdentityRoles(id, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetIdentityRolesResponse{
		Data:   roles.Data,
		Status: 200,
	}

	writeResponse(w, 200, response)
}

// PatchIdentitiesItemRoles Add or remove the identity to/from a role.
// (PATCH /identities/{id}/roles)
func (h handler) PatchIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}
