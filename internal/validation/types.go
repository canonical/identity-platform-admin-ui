// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
)

var (
	NoBodyError = errors.New("request body is not present")
)

func NoMatchError(apiKey string) error {
	return fmt.Errorf("can't find matching validation process for '%s' endpoint", apiKey)
}

func NewValidationError(msg string, errors validator.ValidationErrors) *types.Response {
	return &types.Response{
		Status:  http.StatusBadRequest,
		Message: msg,
		Data:    buildErrorData(errors),
	}
}

func buildErrorData(errors validator.ValidationErrors) map[string][]string {
	if errors == nil {
		return nil
	}

	failedValidations := make(map[string][]string)
	for _, e := range errors {
		field := e.Field()

		failures, ok := failedValidations[field]
		if !ok {
			failedValidations[field] = make([]string, 0)
		}

		failures = append(
			failures,
			fmt.Sprintf("value '%s' fails validation of type `%s`", e.Value(), e.Tag()),
		)
		failedValidations[field] = failures
	}

	return failedValidations
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
