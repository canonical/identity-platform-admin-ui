// Copyright 2024 Canonical Ltd.

package v1

import (
	"encoding/json"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

// writeErrorResponse writes the given err in the response with format defined
// by the OpenAPI spec.
func writeErrorResponse(w http.ResponseWriter, err error) {
	resp := mapErrorResponse(err)

	body, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unexpected marshalling error"))
	}

	w.WriteHeader(int(resp.Status))
	w.Write(body)
}

// mapErrorResponse returns a Response instance filled with the given error.
func mapErrorResponse(err error) resources.Response {
	if isBadRequestError(err) {
		return resources.Response{
			Status:  http.StatusBadRequest,
			Message: "Bad request",
		}
	}

	switch err.(type) {
	case *UnauthorizedError:
		return resources.Response{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	case *NotFoundError:
		return resources.Response{
			Status:  http.StatusNotFound,
			Message: "Not found",
		}
	default:
		return resources.Response{
			Status:  http.StatusInternalServerError,
			Message: "Unexpected error",
		}
	}
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
	}
	return false
}
