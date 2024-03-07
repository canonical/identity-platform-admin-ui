// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetCapabilities returns the list of endpoints implemented by this API.
// (GET /capabilities)
func (h handler) GetCapabilities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	capabilities, err := h.Capabilities.ListCapabilities(ctx)
	if err != nil {
		response := h.CapabilitiesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetCapabilitiesResponse{
		Data:   capabilities.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}
