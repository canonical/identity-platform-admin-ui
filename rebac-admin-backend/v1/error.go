// Copyright 2024 Canonical Ltd.

package v1

// UnauthorizedError represents unauthorized access error.
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "unauthorized"
}

// NotFoundError represents missing entity error.
type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}
