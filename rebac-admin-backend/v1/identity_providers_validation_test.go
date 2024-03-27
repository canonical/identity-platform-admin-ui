// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"go.uber.org/mock/gomock"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

//go:generate mockgen -package resources -destination ./resources/mock_generated_server.go -source=./resources/generated_server.go

func TestHandlerWithValidation_Authentication(t *testing.T) {
	c := qt.New(t)

	writeResponse := func(w http.ResponseWriter, status int, body any) {
		raw, _ := json.Marshal(body)
		w.WriteHeader(status)
		_, _ = w.Write(raw)
	}

	const (
		kindValidationFailure int = 0
		kindSuccessful        int = 1
		kindBadJSON           int = 2
	)

	tests := []struct {
		name             string
		requestBodyRaw   string
		requestBody      any
		setupHandlerMock func(mockHandler *resources.MockServerInterface)
		triggerFunc      func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request)
		kind             int
		expectedPatterns []string
	}{{
		name:        "PostIdentityProviders: success",
		kind:        kindSuccessful,
		requestBody: resources.IdentityProvider{Name: stringPtr("foo"), Id: stringPtr("some-id")},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PostIdentityProviders(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, _ *http.Request) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostIdentityProviders(w, r)
		},
	}, {
		name: "PostIdentityProviders: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostIdentityProviders(w, r)
		},
	}, {
		name:        "PutIdentityProvidersItem: success",
		kind:        kindSuccessful,
		requestBody: resources.IdentityProvider{Name: stringPtr("foo"), Id: stringPtr("some-id")},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PutIdentityProvidersItem(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentityProvidersItem(w, r, "some-id")
		},
	}, {
		name: "PutIdentityProvidersItem: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentityProvidersItem(w, r, "some-id")
		},
	}, {
		name:             "PutIdentityProvidersItem: failure; nil id",
		expectedPatterns: []string{"identity provider ID from path does not match the IdentityProvider object"},
		requestBody:      resources.IdentityProvider{Name: stringPtr("foo")},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentityProvidersItem(w, r, "some-id")
		},
	}, {
		name:             "PutIdentityProvidersItem: failure; id mismatch",
		expectedPatterns: []string{"identity provider ID from path does not match the IdentityProvider object"},
		requestBody:      resources.IdentityProvider{Name: stringPtr("foo"), Id: stringPtr("some-other-id")},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentityProvidersItem(w, r, "some-id")
		},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			ctrl := gomock.NewController(c)
			defer ctrl.Finish()

			mockHandler := resources.NewMockServerInterface(ctrl)
			if tt.setupHandlerMock != nil {
				tt.setupHandlerMock(mockHandler)
			}

			sut := newHandlerWithValidation(mockHandler)

			var req *http.Request
			if tt.requestBody != nil {
				raw, err := json.Marshal(tt.requestBody)
				c.Assert(err, qt.IsNil)
				// Note that request method/URL shouldn't be important at the handler.
				req, _ = http.NewRequest(http.MethodGet, "/blah", bytes.NewReader(raw))
			} else {
				// Note that request method/URL shouldn't be important at the handler.
				req, _ = http.NewRequest(http.MethodGet, "/blah", bytes.NewReader([]byte(tt.requestBodyRaw)))
			}

			mockWriter := httptest.NewRecorder()
			tt.triggerFunc(sut, mockWriter, req)

			response := mockWriter.Result()
			if tt.kind == kindSuccessful {
				c.Assert(response.StatusCode, qt.Equals, http.StatusOK)
			} else {
				c.Assert(response.StatusCode, qt.Equals, http.StatusBadRequest)

				defer response.Body.Close()
				responseBody, err := io.ReadAll(response.Body)
				c.Assert(err, qt.IsNil)

				parsedResponse := &resources.Response{}
				err = json.Unmarshal(responseBody, parsedResponse)
				c.Assert(err, qt.IsNil)
				c.Assert(parsedResponse.Status, qt.Equals, http.StatusBadRequest)

				if tt.kind == kindBadJSON {
					c.Assert(parsedResponse.Message, qt.Matches, "Bad Request: missing request body: request body is not a valid JSON")
				} else if tt.kind == kindValidationFailure {
					c.Assert(parsedResponse.Message, qt.Matches, regexp.MustCompile("Bad Request: invalid request body: .+"))
				}

				for _, pattern := range tt.expectedPatterns {
					c.Assert(parsedResponse.Message, qt.Matches, regexp.MustCompile(fmt.Sprintf(".*%s.*", pattern)))
				}
			}
		})
	}
}
