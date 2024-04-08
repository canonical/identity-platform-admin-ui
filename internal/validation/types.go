// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"net/http"
	"reflect"
	"strings"

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

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// register a function to make 3rd party validator's errors reference json field names instead of Go struct field
	// these errors will be used by frontend code
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}
