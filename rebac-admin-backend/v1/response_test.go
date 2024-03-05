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
			Err:       errors.New("can't find param"),
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Error unmarshaling parameter test-param as JSON: can't find param",
		},
	}, {
		name: "handler error: RequiredParamError",
		arg: &resources.RequiredParamError{
			ParamName: "foo",
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Query argument foo is required, but not found",
		},
	}, {
		name: "handler error: RequiredHeaderError",
		arg: &resources.RequiredHeaderError{
			ParamName: "foo",
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Header parameter foo is required, but not found",
		},
	}, {
		name: "handler error: InvalidParamFormatError",
		arg: &resources.InvalidParamFormatError{
			ParamName: "test-param",
			Err:       errors.New("invalid param"),
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid format for parameter test-param: invalid param",
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
		arg:  NewUnauthorizedError("forbidden"),
		expected: &resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "unauthorized: forbidden",
		},
	}, {
		name: "service error: NotFoundError",
		arg:  NewNotFoundError("test"),
		expected: &resources.Response{
			Status:  http.StatusNotFound,
			Message: "not found: test",
		},
	}, {
		name: "validation error",
		arg:  NewValidationError("request is not valid"),
		expected: &resources.Response{
			Message: "validation error: request is not valid",
			Status:  400,
		},
	}, {
		name: "unknown error",
		arg:  errors.New("unexpected error"),
		expected: &resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "unexpected error",
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
