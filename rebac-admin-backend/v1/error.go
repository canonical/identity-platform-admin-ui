// Copyright 2024 Canonical Ltd.

package v1

import (
	"fmt"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// UnauthorizedError represents unauthorized access error.
type UnauthorizedError struct {
	message string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("Unauthorized: %s", e.message)
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{message}
}

// NotFoundError represents missing entity error.
type NotFoundError struct {
	message string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not found: %s", e.message)
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message}
}

// ErrorResponseMapper is the basic interface to allow for error -> http response mapping
type ErrorResponseMapper interface {
	MapError(error) *resources.Response
}

type delegateErrorResponseMapper struct {
	delegate ErrorResponseMapper
}

func (d delegateErrorResponseMapper) MapError(err error) *resources.Response {
	var response *resources.Response
	if d.delegate != nil {
		response = d.delegate.MapError(err)
	}

	if nil == response {
		response = mapErrorResponse(err)
	}

	return response
}

// NewDefaultErrorResponseMapper returns a pointer to an errorMapper that maps known errors
// to specific error responses and all custom errors to 500 Internal server error responses
func NewDefaultErrorResponseMapper() ErrorResponseMapper {
	return &delegateErrorResponseMapper{}
}

// NewDelegateErrorResponseMapper creates an error mapper that relies on the default error mapping
// only if the provided error mapper returns nil for the error used as input
func NewDelegateErrorResponseMapper(m ErrorResponseMapper) ErrorResponseMapper {
	return &delegateErrorResponseMapper{delegate: m}
}
