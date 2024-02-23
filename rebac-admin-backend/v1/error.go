// Copyright 2024 Canonical Ltd.

package v1

import (
	"fmt"
)

// baseError struct for all error structs
type baseError struct {
	message string
}

// UnauthorizedError represents unauthorized access error.
type UnauthorizedError struct {
	baseError
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("Unauthorized: %s", e.message)
}

// NotFoundError represents missing entity error.
type NotFoundError struct {
	baseError
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not found: %s", e.message)
}
