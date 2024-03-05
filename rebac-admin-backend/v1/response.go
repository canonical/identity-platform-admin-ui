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
func mapErrorResponse(err error) *resources.Response {
	var asErrorWithStatus *errorWithStatus
	if err == nil {
		// Theoretically, this should never happen, but we anyway have to check for
		// a nil argument.
		asErrorWithStatus = &errorWithStatus{status: http.StatusOK}
	} else if e, ok := err.(*errorWithStatus); ok {
		asErrorWithStatus = e
	} else if e := mapHandlerInternalError(err); e != nil {
		asErrorWithStatus = e
	} else {
		asErrorWithStatus = &errorWithStatus{
			status:  http.StatusInternalServerError,
			message: err.Error(),
		}
	}

	return &resources.Response{
		Message: asErrorWithStatus.Error(),
		Status:  asErrorWithStatus.status,
	}
}

// writeResponse is a helper method to avoid verbose repetition of very common instructions
func writeResponse(w http.ResponseWriter, status int, responseObject interface{}) {
	data, err := json.Marshal(responseObject)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	w.WriteHeader(status)

	_, err = w.Write(data)
	if err != nil {
		writeErrorResponse(w, err)
	}
}
