// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import "fmt"

// validateStringEnum is a helper validator to validate string enums.
func validateStringEnum[T ~string](field string, value T, allowed ...T) error {
	for _, element := range allowed {
		if string(value) == string(element) {
			return nil
		}
	}
	return NewRequestBodyValidationError(fmt.Sprintf("%s value not allowed: %q", field, value))
}

// validateStringEnum is a helper validator to validate elements in a slice.
func validateSlice[T any](s []T, validator func(*T) error) error {
	for _, element := range s {
		if err := validator(&element); err != nil {
			return err
		}
	}
	return nil
}
