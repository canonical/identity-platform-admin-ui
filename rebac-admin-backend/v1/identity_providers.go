// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// GetAvailableIdentityProviders returns the list of supported identity providers.
// (GET /authentication/providers)
func (h handler) GetAvailableIdentityProviders(w http.ResponseWriter, req *http.Request, params resources.GetAvailableIdentityProvidersParams) {
	ctx := req.Context()

	identityProviders, err := h.IdentityProviders.ListAvailableIdentityProviders(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	response := resources.GetAvailableIdentityProvidersResponse{
		Links:  resources.NewResponseLinks[resources.AvailableIdentityProvider](req.URL, identityProviders),
		Meta:   identityProviders.Meta,
		Data:   identityProviders.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)

}

// GetIdentityProviders returns a list of registered authentication providers configurations.
// (GET /authentication)
func (h handler) GetIdentityProviders(w http.ResponseWriter, req *http.Request, params resources.GetIdentityProvidersParams) {
	ctx := req.Context()

	identityProviders, err := h.IdentityProviders.ListIdentityProviders(ctx, &params)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	response := resources.GetIdentityProvidersResponse{
		Links:  resources.NewResponseLinks[resources.IdentityProvider](req.URL, identityProviders),
		Meta:   identityProviders.Meta,
		Data:   identityProviders.Data,
		Status: http.StatusOK,
	}

	writeResponse(w, http.StatusOK, response)

}

// PostIdentityProviders register a new authentication provider configuration.
// (POST /authentication)
func (h handler) PostIdentityProviders(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := getRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identityProvider, ok := body.(*resources.IdentityProvider)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.IdentityProviders.RegisterConfiguration(ctx, identityProvider)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusCreated, result)
}

// DeleteIdentityProvidersItem removes an authentication provider configuration identified by `id`.
// (DELETE /authentication/{id})
func (h handler) DeleteIdentityProvidersItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	_, err := h.IdentityProviders.DeleteConfiguration(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetIdentityProvidersItem returns the authentication provider configuration identified by `id`.
// (GET /authentication/{id})
func (h handler) GetIdentityProvidersItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	identityProvider, err := h.IdentityProviders.GetConfiguration(ctx, id)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, identityProvider)
}

// PutIdentityProvidersItem update the authentication provider configuration identified by `id`.
// (PUT /authentication/{id})
func (h handler) PutIdentityProvidersItem(w http.ResponseWriter, req *http.Request, id string) {
	ctx := req.Context()

	body, err := getRequestBodyFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	identityProvider, ok := body.(*resources.IdentityProvider)
	if !ok {
		writeErrorResponse(w, NewMissingRequestBodyError(""))
		return
	}

	result, err := h.IdentityProviders.UpdateConfiguration(ctx, identityProvider)
	if err != nil {
		writeServiceErrorResponse(w, h.IdentityProvidersErrorMapper, err)
		return
	}

	writeResponse(w, http.StatusOK, result)
}
