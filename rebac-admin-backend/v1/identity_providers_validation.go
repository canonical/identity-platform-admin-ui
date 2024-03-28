// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// PostIdentityProviders validates request body for the PostIdentityProviders method and delegates to the underlying handler.
func (v handlerWithValidation) PostIdentityProviders(w http.ResponseWriter, r *http.Request) {
	body := &resources.IdentityProvider{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		v.ServerInterface.PostIdentityProviders(w, r)
	})
}

// PutIdentityProvidersItem validates request body for the PutIdentityProvidersItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutIdentityProvidersItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &resources.IdentityProvider{}
	v.validateRequestBody(body, w, r, func(w http.ResponseWriter, r *http.Request) {
		if body.Id == nil || id != *body.Id {
			writeErrorResponse(w, NewRequestBodyValidationError("identity provider ID from path does not match the IdentityProvider object"))
			return
		}
		v.ServerInterface.PutIdentityProvidersItem(w, r, id)
	})
}
