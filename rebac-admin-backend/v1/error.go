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

// NewUnauthorizedError returns a pointer to a new instance of UnauthorizedError with the provided message
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

// NewNotFoundError returns a pointer to a new instance of NotFoundError with the provided message
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message}
}

// ValidationError represents error in validation of the incoming request
type ValidationError struct {
	message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("Validation error: %s", v.message)
}

// NewValidationError returns a pointer to a new instance of ValidationError with the provided message
func NewValidationError(message string) *ValidationError {
	return &ValidationError{message}
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

// isBadRequestError determines whether the given error should be teated as a
// "Bad Request" (400) error.
func isBadRequestError(err error) bool {
	switch err.(type) {
	case *resources.UnmarshalingParamError:
		return true
	case *resources.RequiredParamError:
		return true
	case *resources.RequiredHeaderError:
		return true
	case *resources.InvalidParamFormatError:
		return true
	case *resources.TooManyValuesForParamError:
		return true
	case *ValidationError:
		return true
	}
	return false
}
