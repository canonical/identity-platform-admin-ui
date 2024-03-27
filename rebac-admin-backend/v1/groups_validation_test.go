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

func TestHandlerWithValidation_Groups(t *testing.T) {
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
		name:        "PostGroups: success",
		kind:        kindSuccessful,
		requestBody: resources.Group{Name: "foo"},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PostGroups(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, _ *http.Request) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostGroups(w, r)
		},
	}, {
		name: "PostGroups: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostGroups(w, r)
		},
	}, {
		name:             "PostGroups: failure; empty",
		expectedPatterns: []string{"'Name' failed on the 'required' tag"},
		requestBody:      resources.Group{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostGroups(w, r)
		},
	}, {
		name:        "PutGroupsItem: success",
		kind:        kindSuccessful,
		requestBody: resources.Group{Name: "foo", Id: stringPtr("some-id")},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PutGroupsItem(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
	}, {
		name: "PutGroupsItem: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
	}, {
		name:             "PutGroupsItem: failure; empty",
		expectedPatterns: []string{"'Name' failed on the 'required' tag"},
		requestBody:      resources.Group{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
	}, {
		name:             "PutGroupsItem: failure; nil id",
		expectedPatterns: []string{"group ID from path does not match the Group object"},
		requestBody:      resources.Group{Name: "foo"},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
	}, {
		name:             "PutGroupsItem: failure; id mismatch",
		expectedPatterns: []string{"group ID from path does not match the Group object"},
		requestBody:      resources.Group{Name: "foo", Id: stringPtr("some-other-id")},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemEntitlements: success",
		kind: kindSuccessful,
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: validEntitlement,
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchGroupsItemEntitlements(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemEntitlements: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.GroupEntitlementsPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemEntitlements: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemEntitlements: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{{
				Op:          "some-invalid-op",
				Entitlement: validEntitlement,
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; invalid entitlement",
		expectedPatterns: []string{
			"'EntitlementType' failed on the 'required' tag",
			"'EntityName' failed on the 'required' tag",
			"'EntityType' failed on the 'required' tag",
		},
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: resources.EntityEntitlement{},
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemIdentities: success",
		kind: kindSuccessful,
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{{
				Op:       "add",
				Identity: "some-identity",
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchGroupsItemIdentities(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemIdentities: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemIdentities: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.GroupIdentitiesPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemIdentities: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemIdentities: failure; empty identity",
		expectedPatterns: []string{"'Identity' failed on the 'required' tag"},
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{{
				Op:       "add",
				Identity: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemIdentities: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{{
				Op:       "some-invalid-op",
				Identity: "some-identity",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemRoles: success",
		kind: kindSuccessful,
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{{
				Op:   "add",
				Role: "some-role",
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchGroupsItemRoles(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
	}, {
		name: "PatchGroupsItemRoles: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemRoles: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.GroupRolesPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemRoles: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemRoles: failure; empty role",
		expectedPatterns: []string{"'Role' failed on the 'required' tag"},
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{{
				Op:   "add",
				Role: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchGroupsItemRoles: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{{
				Op:   "some-invalid-op",
				Role: "some-role",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
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
