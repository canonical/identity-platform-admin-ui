// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package kratos

import (
	"encoding/json"
	"io"
	"net/http"

	kClient "github.com/ory/kratos-client-go"
)

// TODO @shipperizer verify during integration test if this is actually the format
type KratosError struct {
	Error *kClient.GenericError `json:"error,omitempty"`
}

func ParseKratosError(r *http.Response) *kClient.GenericError {
	gerr := KratosError{Error: kClient.NewGenericErrorWithDefaults()}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(body, &gerr); err != nil {
		gerr.Error.SetMessage("unable to parse kratos error response")
		gerr.Error.SetCode(http.StatusInternalServerError)
	}

	return gerr.Error
}
