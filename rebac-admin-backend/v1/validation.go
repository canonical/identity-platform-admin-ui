// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// handlerWithValidation decorates a given handler with validation logic. The
// request body is parsed into a safely-typed value and passed to the handler
// via context.
type handlerWithValidation struct {
	// Wrapped/decorated handler
	resources.ServerInterface

	validate *validator.Validate
}

// newHandlerWithValidation returns a new instance of the validationHandlerDecorator struct.
func newHandlerWithValidation(handler resources.ServerInterface) *handlerWithValidation {
	return &handlerWithValidation{
		ServerInterface: handler,
		validate:        validator.New(),
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

// validateRequestBody is a helper method to avoid repetition. It parses
// request body, validates it against the given validator instance and if it's
// okay, will delegate to the provided callback with a new HTTP request instance
// with the parse body in the context.
//
// Note that, technically, this method could be an ordinary (non-generic) method,
// but it's defined as one to avoid confusion over value vs pointer arguments.
func validateRequestBody[T any](v *validator.Validate, w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter, r *http.Request, body *T)) {
	body, err := parseRequestBody[T](r)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	if err := v.Struct(body); err != nil {
		writeErrorResponse(w, NewRequestBodyValidationError(err.Error()))
		return
	}
	f(w, newRequestWithBodyInContext(r, body), body)
}
