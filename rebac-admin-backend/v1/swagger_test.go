// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestHandler_SwaggerJson(t *testing.T) {
	c := qt.New(t)

	sut := handler{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger.json", nil)
	sut.SwaggerJson(w, req)

	result := w.Result()
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	parsedSpec := map[string]any{}
	err = json.Unmarshal(body, &parsedSpec)
	c.Assert(err, qt.IsNil)
	c.Assert(len(parsedSpec) > 0, qt.IsTrue)
}
