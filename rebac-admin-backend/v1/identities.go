// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"net/http"

	r "github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetIdentities Get list of identities.
// (GET /identities)
func (h handler) GetIdentities(w http.ResponseWriter, req *http.Request, params r.GetIdentitiesParams) {
	identities, err := h.Identities.ListIdentities(&params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	response := r.GetIdentitiesResponse{
		Data:   identities.Data,
		Status: 200,
	}

	marshalled, err := json.Marshal(response)

	writeResponse(w, 200, marshalled, err)
}

// PostIdentities Add an identity.
// (POST /identities)
func (h handler) PostIdentities(w http.ResponseWriter, req *http.Request) {
	identity := &r.Identity{}
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(identity)
	if err != nil {
		writeResponse(w, 500, nil, err)
		return
	}

	identity, err = h.Identities.CreateIdentity(identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	marshalled, err := json.Marshal(identity)
	writeResponse(w, 201, marshalled, err)
}

// DeleteIdentitiesItem Remove an identity.
// (DELETE /identities/{id})
func (h handler) DeleteIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	_, err := h.Identities.DeleteIdentity(id)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	w.WriteHeader(200)
}

// GetIdentitiesItem Get a single identity.
// (GET /identities/{id})
func (h handler) GetIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	identity, err := h.Identities.GetIdentity(id)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	marshalled, err := json.Marshal(identity)
	writeResponse(w, 200, marshalled, err)
}

// PutIdentitiesItem Update an identity.
// (PUT /identities/{id})
func (h handler) PutIdentitiesItem(w http.ResponseWriter, req *http.Request, id string) {
	identity := &r.Identity{}
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(identity)
	if err != nil {
		writeResponse(w, 500, nil, err)
		return
	}

	identity, err = h.Identities.UpdateIdentity(identity)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	marshalled, err := json.Marshal(identity)
	writeResponse(w, 201, marshalled, err)
}

// GetIdentitiesItemEntitlements List entitlements the identity has.
// (GET /identities/{id}/entitlements)
func (h handler) GetIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string, params r.GetIdentitiesItemEntitlementsParams) {
	var identity *r.Identity

	err := json.NewDecoder(req.Body).Decode(identity)
	if err != nil {
		writeResponse(w, 500, nil, err)
		return
	}

	//entitlements, err := h.Identities.GetIdentityEntitlements(id, &params)
	if err != nil {
		response := h.IdentitiesErrorMapper.MapError(err)
		marshal, err := json.Marshal(response)

		writeResponse(w, response.Status, marshal, err)
		return
	}

	//response := r.EntityEntitlements
	/*response := r.GetIdentitiesItemEntitlements{
		Data:   entitlements.Data,
		Status: 200,
	}

	marshalled, err := json.Marshal(response)
	writeResponse(w, response.Status, marshalled, err)*/
}

// PatchIdentitiesItemEntitlements Add or remove entitlement to/from an identity.
// (PATCH /identities/{id}/entitlements)
func (h handler) PatchIdentitiesItemEntitlements(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetIdentitiesItemGroups List groups the identity is a member of.
// (GET /identities/{id}/groups)
func (h handler) GetIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string, params r.GetIdentitiesItemGroupsParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PatchIdentitiesItemGroups Add or remove the identity to/from a group.
// (PATCH /identities/{id}/groups)
func (h handler) PatchIdentitiesItemGroups(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetIdentitiesItemRoles List roles assigned to the identity.
// (GET /identities/{id}/roles)
func (h handler) GetIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string, params r.GetIdentitiesItemRolesParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PatchIdentitiesItemRoles Add or remove the identity to/from a role.
// (PATCH /identities/{id}/roles)
func (h handler) PatchIdentitiesItemRoles(w http.ResponseWriter, req *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}
