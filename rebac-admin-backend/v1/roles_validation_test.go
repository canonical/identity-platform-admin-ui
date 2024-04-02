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

	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

//go:generate mockgen -package resources -destination ./resources/mock_generated_server.go -source=./resources/generated_server.go

func TestHandlerWithValidation_Roles(t *testing.T) {
	c := qt.New(t)

	writeResponse := func(w http.ResponseWriter, status int, body any) {
		raw, _ := json.Marshal(body)
		w.WriteHeader(status)
		_, _ = w.Write(raw)
	}

	validEntitlement := resources.EntityEntitlement{
		EntitlementType: "some-entitlement-type",
		EntityName:      "some-entity-name",
		EntityType:      "some-entity-type",
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
		name:        "PostRoles: success",
		kind:        kindSuccessful,
		requestBody: resources.Role{Name: "foo"},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PostRoles(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, _ *http.Request) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostRoles(w, r)
		},
	}, {
		name: "PostRoles: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostRoles(w, r)
		},
	}, {
		name:             "PostRoles: failure; empty",
		expectedPatterns: []string{"'Name' failed on the 'required' tag"},
		requestBody:      resources.Role{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostRoles(w, r)
		},
	}, {
		name:        "PutRolesItem: success",
		kind:        kindSuccessful,
		requestBody: resources.Role{Name: "foo", Id: stringPtr("some-id")},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PutRolesItem(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutRolesItem(w, r, "some-id")
		},
	}, {
		name: "PutRolesItem: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutRolesItem(w, r, "some-id")
		},
	}, {
		name:             "PutRolesItem: failure; empty",
		expectedPatterns: []string{"'Name' failed on the 'required' tag"},
		requestBody:      resources.Role{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutRolesItem(w, r, "some-id")
		},
	}, {
		name:             "PutRolesItem: failure; nil id",
		expectedPatterns: []string{"role ID from path does not match the Role object"},
		requestBody:      resources.Role{Name: "foo"},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutRolesItem(w, r, "some-id")
		},
	}, {
		name:             "PutRolesItem: failure; id mismatch",
		expectedPatterns: []string{"role ID from path does not match the Role object"},
		requestBody:      resources.Role{Name: "foo", Id: stringPtr("some-other-id")},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutRolesItem(w, r, "some-id")
		},
	}, {
		name: "PatchRolesItemEntitlements: success",
		kind: kindSuccessful,
		requestBody: resources.RoleEntitlementsPatchRequestBody{
			Patches: []resources.RoleEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: validEntitlement,
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchRolesItemEntitlements(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchRolesItemEntitlements: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchRolesItemEntitlements: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.RoleEntitlementsPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchRolesItemEntitlements: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.RoleEntitlementsPatchRequestBody{
			Patches: []resources.RoleEntitlementsPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchRolesItemEntitlements: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.RoleEntitlementsPatchRequestBody{
			Patches: []resources.RoleEntitlementsPatchItem{{
				Op:          "some-invalid-op",
				Entitlement: validEntitlement,
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchRolesItemEntitlements: failure; invalid entitlement",
		expectedPatterns: []string{
			"'EntitlementType' failed on the 'required' tag",
			"'EntityName' failed on the 'required' tag",
			"'EntityType' failed on the 'required' tag",
		},
		requestBody: resources.RoleEntitlementsPatchRequestBody{
			Patches: []resources.RoleEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: resources.EntityEntitlement{},
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchRolesItemEntitlements(w, r, "some-id")
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
