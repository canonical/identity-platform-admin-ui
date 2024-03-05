// Copyright 2024 Canonical Ltd.

package v1

import (
	"fmt"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// errorWithStatus is an internal error representation that holds the corresponding
// HTTP status code along with the error message.
type errorWithStatus struct {
	// status is the HTTP standard equivalent status. Acceptable
	// values are `http.Status*` constants.
	status  int
	message string
}

// Error implements the error interface.
func (e *errorWithStatus) Error() string {
	statusText := http.StatusText(e.status)
	if statusText == "" {
		statusText = "[Unknown error]"
	}
	if e.message == "" {
		return statusText
	}
	return fmt.Sprintf("%s: %s", statusText, e.message)
}

// NewUnauthorizedError returns an error instance that represents an unauthorized access error.
func NewUnauthorizedError(message string) error {
	return &errorWithStatus{
		status:  http.StatusUnauthorized,
		message: message,
	}
}

// NewNotFoundError returns an error instance that represents a not-found error.
func NewNotFoundError(message string) error {
	return &errorWithStatus{
		status:  http.StatusNotFound,
		message: message,
	}
}

// NewValidationError returns an error instance that represents an input validation error.
func NewValidationError(message string) error {
	return &errorWithStatus{
		status:  http.StatusBadRequest,
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

// mapHandlerBadRequestError checks if the given error is an "Bad Request" error
// thrown at the handler root (i.e., an auto-generated error type) and return the
// equivalent errorWithStatus instance. If the given error is not an internal
// handler error, this function will return nil.
func mapHandlerBadRequestError(err error) *errorWithStatus {
	if !isHandlerBadRequestError(err) {
		return nil
	}
	return &errorWithStatus{
		status:  http.StatusBadRequest,
		message: err.Error(),
	}
}

func isHandlerBadRequestError(err error) bool {
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
