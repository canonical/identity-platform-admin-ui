// Copyright 2024 Canonical Ltd.

package v1

import (
	"errors"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

func TestMapErrorResponse(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		name     string
		arg      error
		expected *resources.Response
	}{{
		name: "handler error: UnmarshalingParamError",
		arg: &resources.UnmarshalingParamError{
			ParamName: "test-param",
			Err:       errors.New("Can't find param"),
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Error unmarshaling parameter test-param as JSON: Can't find param",
		},
	}, {
		name: "handler error: RequiredParamError",
		arg:  &resources.RequiredParamError{},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Query argument  is required, but not found",
		},
	}, {
		name: "handler error: RequiredHeaderError",
		arg:  &resources.RequiredHeaderError{},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Header parameter  is required, but not found",
		},
	}, {
		name: "handler error: InvalidParamFormatError",
		arg: &resources.InvalidParamFormatError{
			ParamName: "test-param",
			Err:       errors.New("Invalid param"),
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid format for parameter test-param: Invalid param",
		},
	}, {
		name: "handler error: TooManyValuesForParamError",
		arg:  &resources.TooManyValuesForParamError{ParamName: "test-param"},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Expected one value for test-param, got 0",
		},
	}, {
		name: "service error: UnauthorizedError",
		arg:  &UnauthorizedError{"forbidden"},
		expected: &resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized: forbidden",
		},
	}, {
		name: "service error: NotFoundError",
		arg:  &NotFoundError{"test"},
		expected: &resources.Response{
			Status:  http.StatusNotFound,
			Message: "Not found: test",
		},
	}, {
		name: "validation error",
		arg:  &ValidationError{message: "request is not valid"},
		expected: &resources.Response{
			Message: "Validation error: request is not valid",
			Status:  400,
		},
	}, {
		name: "unknown error",
		arg:  errors.New("Unexpected error"),
		expected: &resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "Unexpected error",
		},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			value := mapErrorResponse(tt.arg)
			c.Assert(value, qt.DeepEquals, tt.expected)
		})
	}
}
