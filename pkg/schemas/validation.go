// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	identitySchemaContainerRules = map[string]string{
		"Schema": "required",
	}
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate
}

func isPartialUpdate(r *http.Request, endpoint string) bool {
	return strings.HasPrefix(endpoint, "/") && r.Method == http.MethodPatch
}

func isCreateOrUpdateSchema(r *http.Request, endpoint string) bool {
	return (endpoint == "" && r.Method == http.MethodPost) || (endpoint == "/default" && r.Method == http.MethodPut)
}

func shouldValidate(r *http.Request) bool {
	return r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch
}
