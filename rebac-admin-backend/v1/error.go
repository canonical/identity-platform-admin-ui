// Copyright 2024 Canonical Ltd.

package v1

import (
	"fmt"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// ErrorCode is a code which describes the class of error.
type ErrorCode string

const (
	CodeUnauthorized    ErrorCode = "unauthorized"
	CodeNotFound        ErrorCode = "not found"
	CodeValidationError ErrorCode = "validation error"
)

type errorWithCode struct {
	code    ErrorCode
	message string
}

// Error implements the error interface.
func (e *errorWithCode) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// NewUnauthorizedError returns an error instance that represents an unauthorized access error.
func NewUnauthorizedError(message string) error {
	return &errorWithCode{
		code:    CodeUnauthorized,
		message: message,
	}
}

// NewNotFoundError returns an error instance that represents a not-found error.
func NewNotFoundError(message string) error {
	return &errorWithCode{
		code:    CodeNotFound,
		message: message,
	}
}

// NewValidationError returns an error instance that represents an input validation error.
func NewValidationError(message string) error {
	return &errorWithCode{
		code:    CodeValidationError,
		message: message,
	}
}

// ErrorResponseMapper is the basic interface to allow for error -> http response mapping
type ErrorResponseMapper interface {
	// MapError maps an error into a Response. If the method is unable to map the
	// error (e.g., the error is unknown), it must return nil.
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
	// Note that these are all auto-generated error types, that should be
	// regarded as bad request errors.
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
	}
	return false
}
