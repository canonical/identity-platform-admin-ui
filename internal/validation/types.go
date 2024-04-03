// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
)

func NewValidationError(errors validator.ValidationErrors) *types.Response {
	return &types.Response{
		Status:  http.StatusBadRequest,
		Message: "validation errors",
		Data:    buildErrorData(errors),
	}
}

func buildErrorData(err validator.ValidationErrors) []any {
	return nil
}
