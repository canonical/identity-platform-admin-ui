// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// handlerWithValidation decorates a given handler with validation logic. The
// request body is parsed into a safely-typed value and passed to the handler
// via context.
type handlerWithValidation struct {
	// TODO(babakks) remove the embedded struct.
	resources.Unimplemented

	handler resources.ServerInterface
}

// newHandlerWithValidation returns a new instance of the validationHandlerDecorator struct.
func newHandlerWithValidation(handler resources.ServerInterface) *handlerWithValidation {
	return &handlerWithValidation{
		handler: handler,
	}
}

// requestBodyContextKey is the context key to retrieve the parsed request body struct instance.
type requestBodyContextKey struct{}

// getRequestBodyFromContext fetches request body from given context.
func getRequestBodyFromContext[T any](ctx context.Context) (*T, error) {
	if body, ok := ctx.Value(requestBodyContextKey{}).(*T); ok {
		return body, nil
	}
	return nil, NewMissingRequestBodyError("request body is not available")
}

// newRequestWithBodyInContext sets the given body in a new request instance context
// and returns the new request.
//
// Note that, technically, this method could be an ordinary (non-generic) method,
// but it's defined as one to avoid confusion over value vs pointer arguments.
func newRequestWithBodyInContext[T any](r *http.Request, body *T) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestBodyContextKey{}, body))
}

// parseRequestBody parses request body as JSON and populates the given body instance.
func parseRequestBody[T any](r *http.Request) (*T, error) {
	body := new(T)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		return nil, NewMissingRequestBodyError("request body is not a valid JSON")
	}
	return body, nil
}

// setRequestBodyInContext is a helper method to avoid repetition. It parses
// request body and if it's okay, will delegate to the provided callback with a
// new HTTP request instance with the parse body in the context.
func setRequestBodyInContext[T any](w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter, r *http.Request, body *T)) {
	body, err := parseRequestBody[T](r)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	f(w, newRequestWithBodyInContext(r, body), body)
}

// GetIdentityProviders validates request body for the GetIdentityProviders method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentityProviders(w http.ResponseWriter, r *http.Request, params resources.GetIdentityProvidersParams) {
	v.handler.GetIdentityProviders(w, r, params)
}

// PostIdentityProviders validates request body for the PostIdentityProviders method and delegates to the underlying handler.
func (v handlerWithValidation) PostIdentityProviders(w http.ResponseWriter, r *http.Request) {
	body := &struct{}{}
	v.handler.PostIdentityProviders(w, newRequestWithBodyInContext(r, body))
}

// GetAvailableIdentityProviders validates request body for the GetAvailableIdentityProviders method and delegates to the underlying handler.
func (v handlerWithValidation) GetAvailableIdentityProviders(w http.ResponseWriter, r *http.Request, params resources.GetAvailableIdentityProvidersParams) {
	body := &struct{}{}
	v.handler.GetAvailableIdentityProviders(w, newRequestWithBodyInContext(r, body), params)
}

// DeleteIdentityProvidersItem validates request body for the DeleteIdentityProvidersItem method and delegates to the underlying handler.
func (v handlerWithValidation) DeleteIdentityProvidersItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.DeleteIdentityProvidersItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetIdentityProvidersItem validates request body for the GetIdentityProvidersItem method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentityProvidersItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.GetIdentityProvidersItem(w, newRequestWithBodyInContext(r, body), id)
}

// PutIdentityProvidersItem validates request body for the PutIdentityProvidersItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutIdentityProvidersItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PutIdentityProvidersItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetCapabilities validates request body for the GetCapabilities method and delegates to the underlying handler.
func (v handlerWithValidation) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	body := &struct{}{}
	v.handler.GetCapabilities(w, newRequestWithBodyInContext(r, body))
}

// GetEntitlements validates request body for the GetEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) GetEntitlements(w http.ResponseWriter, r *http.Request, params resources.GetEntitlementsParams) {
	body := &struct{}{}
	v.handler.GetEntitlements(w, newRequestWithBodyInContext(r, body), params)
}

// GetRawEntitlements validates request body for the GetRawEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) GetRawEntitlements(w http.ResponseWriter, r *http.Request) {
	body := &struct{}{}
	v.handler.GetRawEntitlements(w, newRequestWithBodyInContext(r, body))
}

// GetIdentities validates request body for the GetIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentities(w http.ResponseWriter, r *http.Request, params resources.GetIdentitiesParams) {
	body := &struct{}{}
	v.handler.GetIdentities(w, newRequestWithBodyInContext(r, body), params)
}

// PostIdentities validates request body for the PostIdentities method and delegates to the underlying handler.
func (v handlerWithValidation) PostIdentities(w http.ResponseWriter, r *http.Request) {
	body := &struct{}{}
	v.handler.PostIdentities(w, newRequestWithBodyInContext(r, body))
}

// DeleteIdentitiesItem validates request body for the DeleteIdentitiesItem method and delegates to the underlying handler.
func (v handlerWithValidation) DeleteIdentitiesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.DeleteIdentitiesItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetIdentitiesItem validates request body for the GetIdentitiesItem method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentitiesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.GetIdentitiesItem(w, newRequestWithBodyInContext(r, body), id)
}

// PutIdentitiesItem validates request body for the PutIdentitiesItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutIdentitiesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PutIdentitiesItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetIdentitiesItemEntitlements validates request body for the GetIdentitiesItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentitiesItemEntitlements(w http.ResponseWriter, r *http.Request, id string, params resources.GetIdentitiesItemEntitlementsParams) {
	body := &struct{}{}
	v.handler.GetIdentitiesItemEntitlements(w, newRequestWithBodyInContext(r, body), id, params)
}

// PatchIdentitiesItemEntitlements validates request body for the PatchIdentitiesItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchIdentitiesItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PatchIdentitiesItemEntitlements(w, newRequestWithBodyInContext(r, body), id)
}

// GetIdentitiesItemGroups validates request body for the GetIdentitiesItemGroups method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentitiesItemGroups(w http.ResponseWriter, r *http.Request, id string, params resources.GetIdentitiesItemGroupsParams) {
	body := &struct{}{}
	v.handler.GetIdentitiesItemGroups(w, newRequestWithBodyInContext(r, body), id, params)
}

// PatchIdentitiesItemGroups validates request body for the PatchIdentitiesItemGroups method and delegates to the underlying handler.
func (v handlerWithValidation) PatchIdentitiesItemGroups(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PatchIdentitiesItemGroups(w, newRequestWithBodyInContext(r, body), id)
}

// GetIdentitiesItemRoles validates request body for the GetIdentitiesItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) GetIdentitiesItemRoles(w http.ResponseWriter, r *http.Request, id string, params resources.GetIdentitiesItemRolesParams) {
	body := &struct{}{}
	v.handler.GetIdentitiesItemRoles(w, newRequestWithBodyInContext(r, body), id, params)
}

// PatchIdentitiesItemRoles validates request body for the PatchIdentitiesItemRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PatchIdentitiesItemRoles(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PatchIdentitiesItemRoles(w, newRequestWithBodyInContext(r, body), id)
}

// GetResources validates request body for the GetResources method and delegates to the underlying handler.
func (v handlerWithValidation) GetResources(w http.ResponseWriter, r *http.Request, params resources.GetResourcesParams) {
	body := &struct{}{}
	v.handler.GetResources(w, newRequestWithBodyInContext(r, body), params)
}

// GetRoles validates request body for the GetRoles method and delegates to the underlying handler.
func (v handlerWithValidation) GetRoles(w http.ResponseWriter, r *http.Request, params resources.GetRolesParams) {
	body := &struct{}{}
	v.handler.GetRoles(w, newRequestWithBodyInContext(r, body), params)
}

// PostRoles validates request body for the PostRoles method and delegates to the underlying handler.
func (v handlerWithValidation) PostRoles(w http.ResponseWriter, r *http.Request) {
	body := &struct{}{}
	v.handler.PostRoles(w, newRequestWithBodyInContext(r, body))
}

// DeleteRolesItem validates request body for the DeleteRolesItem method and delegates to the underlying handler.
func (v handlerWithValidation) DeleteRolesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.DeleteRolesItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetRolesItem validates request body for the GetRolesItem method and delegates to the underlying handler.
func (v handlerWithValidation) GetRolesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.GetRolesItem(w, newRequestWithBodyInContext(r, body), id)
}

// PutRolesItem validates request body for the PutRolesItem method and delegates to the underlying handler.
func (v handlerWithValidation) PutRolesItem(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PutRolesItem(w, newRequestWithBodyInContext(r, body), id)
}

// GetRolesItemEntitlements validates request body for the GetRolesItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) GetRolesItemEntitlements(w http.ResponseWriter, r *http.Request, id string, params resources.GetRolesItemEntitlementsParams) {
	body := &struct{}{}
	v.handler.GetRolesItemEntitlements(w, newRequestWithBodyInContext(r, body), id, params)
}

// PatchRolesItemEntitlements validates request body for the PatchRolesItemEntitlements method and delegates to the underlying handler.
func (v handlerWithValidation) PatchRolesItemEntitlements(w http.ResponseWriter, r *http.Request, id string) {
	body := &struct{}{}
	v.handler.PatchRolesItemEntitlements(w, newRequestWithBodyInContext(r, body), id)
}

// SwaggerJson validates request body for the SwaggerJson method and delegates to the underlying handler.
func (v handlerWithValidation) SwaggerJson(w http.ResponseWriter, r *http.Request) {
	v.handler.SwaggerJson(w, r)
}
