// Copyright 2024 Canonical Ltd.

package v1

import (
	"errors"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

func TestGetErrorResponse(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		name     string
		arg      error
		expected resources.Response
	}{{
		name: "handler error: UnmarshalingParamError",
		arg:  &resources.UnmarshalingParamError{},
		expected: resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		},
	}, {
		name: "handler error: RequiredParamError",
		arg:  &resources.RequiredParamError{},
		expected: resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		},
	}, {
		name: "handler error: RequiredHeaderError",
		arg:  &resources.RequiredHeaderError{},
		expected: resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		},
	}, {
		name: "handler error: InvalidParamFormatError",
		arg:  &resources.InvalidParamFormatError{},
		expected: resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		},
	}, {
		name: "handler error: TooManyValuesForParamError",
		arg:  &resources.TooManyValuesForParamError{},
		expected: resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		},
	}, {
		name: "service error: UnauthorizedError",
		arg:  &UnauthorizedError{},
		expected: resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized",
		},
	}, {
		name: "service error: NotFoundError",
		arg:  &NotFoundError{},
		expected: resources.Response{
			Status:  http.StatusNotFound,
			Message: "Not found",
		},
	}, {
		name: "unknown error",
		arg:  errors.New("something went wrong"),
		expected: resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "Unexpected error",
		},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			value := getErrorResponse(tt.arg)
			c.Assert(value, qt.DeepEquals, tt.expected)
		})
	}
}
