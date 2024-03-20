// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"net/http"
	"net/http/httptest"
)

func newTestRequest[T any](method, path string, body *T) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	return newRequestWithBodyInContext(r, body)
}

func stringPtr(s string) *string {
	return &s
}
