// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"
	"net/http/httptest"
)

// newTestRequest returns a new HTTP request instance with the given body set in the
// corresponding context.
func newTestRequest[T any](method, path string, body *T) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	return newRequestWithBodyInContext(r, body)
}

// stringPtr is a helper function that returns a pointer to the given string literal.
func stringPtr(s string) *string {
	return &s
}
