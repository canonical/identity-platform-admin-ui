// Copyright 2024 Canonical Ltd.

package v1

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

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
			Message: "Bad Request: Error unmarshaling parameter test-param as JSON: can't find param",
		},
	}, {
		name: "handler error: RequiredParamError",
		arg: &resources.RequiredParamError{
			ParamName: "foo",
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad Request: Query argument foo is required, but not found",
		},
	}, {
		name: "handler error: RequiredHeaderError",
		arg: &resources.RequiredHeaderError{
			ParamName: "foo",
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad Request: Header parameter foo is required, but not found",
		},
	}, {
		name: "handler error: InvalidParamFormatError",
		arg: &resources.InvalidParamFormatError{
			ParamName: "test-param",
			Err:       errors.New("invalid param"),
		},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad Request: Invalid format for parameter test-param: invalid param",
		},
	}, {
		name: "handler error: TooManyValuesForParamError",
		arg:  &resources.TooManyValuesForParamError{ParamName: "test-param"},
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad Request: Expected one value for test-param, got 0",
		},
	}, {
		name: "service error: AuthenticationError",
		arg:  NewAuthenticationError("unknown user"),
		expected: &resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized: authentication failed: unknown user",
		},
	}, {
		name: "service error: UnauthorizedError",
		arg:  NewAuthorizationError("forbidden"),
		expected: &resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized: authorization failed: forbidden",
		},
	}, {
		name: "service error: NotFoundError",
		arg:  NewNotFoundError("something not found"),
		expected: &resources.Response{
			Status:  http.StatusNotFound,
			Message: "Not Found: something not found",
		},
	}, {
		name: "validation error",
		arg:  NewValidationError("request is not valid"),
		expected: &resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad Request: request is not valid",
		},
	}, {
		name: "unknown error",
		arg:  errors.New("unexpected error"),
		expected: &resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "Internal Server Error: unexpected error",
		},
	}, {
		name: "nil error",
		arg:  nil,
		expected: &resources.Response{
			Status:  http.StatusOK,
			Message: "OK",
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

func TestMapServiceErrorResponse(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		initMapper func() ErrorResponseMapper
		err        error
		expected   resources.Response
	}{{
		name: "nil mapper",
		err:  errors.New("foo"),
		expected: resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "Internal Server Error: foo",
		},
	}, {
		name: "non-nil mapper",
		initMapper: func() ErrorResponseMapper {
			mapper := NewMockErrorResponseMapper(ctrl)
			mapper.EXPECT().
				MapError(gomock.Any()).
				Return(&resources.Response{
					Status:  999, // Some bizarre status code
					Message: "foo",
				})
			return mapper
		},
		err: errors.New("bar"),
		expected: resources.Response{
			Status:  999,
			Message: "foo",
		},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			var mapper ErrorResponseMapper
			if tt.initMapper != nil {
				mapper = tt.initMapper()
			}

			response := mapServiceErrorResponse(mapper, tt.err)
			c.Assert(*response, qt.DeepEquals, tt.expected)
		})
	}
}
