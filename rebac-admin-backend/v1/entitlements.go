// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetEntitlements returns the list of known entitlements in JSON format.
// (GET /entitlements)
func (h handler) GetEntitlements(w http.ResponseWriter, req *http.Request, params resources.GetEntitlementsParams) {
	ctx := req.Context()

	entitlements, err := h.Entitlements.ListEntitlements(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.EntitlementsErrorMapper, err)
		return
	}

	response := resources.GetEntitlementsResponse{
		Links:  resources.NewResponseLinks[resources.EntityEntitlement](req.URL, entitlements),
		Meta:   entitlements.Meta,
		Data:   entitlements.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)

}

// GetRawEntitlements returns the list of known entitlements as raw text.
// (GET /entitlements/raw)
func (h handler) GetRawEntitlements(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	entitlementsRawString, err := h.Entitlements.RawEntitlements(ctx)
	if err != nil {
		writeServiceErrorResponse(w, h.EntitlementsErrorMapper, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	writeResponse(w, http.StatusOK, entitlementsRawString)

}