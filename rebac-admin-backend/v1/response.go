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
	if isBadRequestError(err) {
		return &resources.Response{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	var status int

	switch err.(type) {
	case *UnauthorizedError:
		status = http.StatusUnauthorized
	case *NotFoundError:
		status = http.StatusNotFound
	default:
		status = http.StatusInternalServerError
	}

	return &resources.Response{
		Message: err.Error(),
		Status:  status,
	}
}

// writeResponse is a helper method to avoid verbose repetition of very common instructions
// it writes to the ResponseWriter either:
// - the provided marshalledResponse if marshalErr is nil
// - or the mapped error response based on marshalErr if it is not nil
func writeResponse(w http.ResponseWriter, status int, marshalledResp []byte, marshalErr error) {
	if marshalErr != nil {
		writeErrorResponse(w, marshalErr)
		return
	}

	w.WriteHeader(status)
	_, err := w.Write(marshalledResp)

	if err != nil {
		writeErrorResponse(w, err)
	}
}
