// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"log"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// Returns the OpenAPI spec as a JSON file.
// (GET /swagger.json)
func (h handler) SwaggerJson(w http.ResponseWriter, req *http.Request) {
	swagger, err := resources.GetSwagger()
	if err != nil {
		writeErrorResponse(w, NewUnknownError("cannot retrieve swagger data"))
		return
	}

	body, err := swagger.MarshalJSON()
	if err != nil {
		writeErrorResponse(w, NewUnknownError("cannot marshal spec as JSON"))
		return
	}

	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write response body: %v", err)
	}
}
