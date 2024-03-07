// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetResources returns the list of known resources.
// (GET /resources)
func (h handler) GetResources(w http.ResponseWriter, req *http.Request, params resources.GetResourcesParams) {
	ctx := req.Context()

	res, err := h.Resources.ListResources(ctx, &params)
	if err != nil {
		response := h.ResourcesErrorMapper.MapError(err)
		writeResponse(w, response.Status, response)
		return
	}

	response := resources.GetResourcesResponse{
		Data:   res.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)
}
