// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -package resources -destination ./resources/mock_generated_server.go -source=./resources/generated_server.go

func TestHandlerWithValidation_Groups(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writeResponse := func(w http.ResponseWriter, status int, body any) {
		raw, _ := json.Marshal(body)
		w.WriteHeader(status)
		w.Write(raw)
	}

	validEntitlement := resources.EntityEntitlement{
		EntitlementType: "some-entitlement-type",
		EntityName:      "some-entity-name",
		EntityType:      "some-entity-type",
	}

	tests := []struct {
		name               string
		requestBodyRaw     string
		requestBody        any
		setupHandlerMock   func(mockHandler *resources.MockServerInterface)
		triggerFunc        func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request)
		expectedStatusCode int
		expectedResponse   any
	}{{
		name:        "PostGroups: success",
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
		expectedStatusCode: http.StatusOK,
		expectedResponse: resources.Response{
			Status: http.StatusOK,
		},
	}, {
		name: "PostGroups: failure; invalid JSON",
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostGroups(w, r)
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: missing request body: request body is not a valid JSON",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PostGroups: failure; empty",
		requestBody: resources.Group{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostGroups(w, r)
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty group name",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PutGroupsItem: success",
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
		expectedStatusCode: http.StatusOK,
		expectedResponse: resources.Response{
			Status: http.StatusOK,
		},
	}, {
		name: "PutGroupsItem: failure; invalid JSON",
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: missing request body: request body is not a valid JSON",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PutGroupsItem: failure; empty",
		requestBody: resources.Group{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty group name",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PutGroupsItem: failure; nil id",
		requestBody: resources.Group{Name: "foo"},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: group ID from path does not match the Group object",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PutGroupsItem: failure; id mismatch",
		requestBody: resources.Group{Name: "foo", Id: stringPtr("some-other-id")},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutGroupsItem(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: group ID from path does not match the Group object",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemEntitlements: success",
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
		expectedStatusCode: http.StatusOK,
		expectedResponse: resources.Response{
			Status: http.StatusOK,
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; invalid JSON",
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: missing request body: request body is not a valid JSON",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PatchGroupsItemEntitlements: failure; nil patch array",
		requestBody: resources.GroupEntitlementsPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; empty patch array",
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; invalid op",
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{{
				Op:          "some-invalid-op",
				Entitlement: validEntitlement,
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: op value not allowed: \"some-invalid-op\"",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemEntitlements: failure; invalid entitlement",
		requestBody: resources.GroupEntitlementsPatchRequestBody{
			Patches: []resources.GroupEntitlementsPatchItem{{
				Op:          "some-invalid-op",
				Entitlement: resources.EntityEntitlement{},
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemEntitlements(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty entitlement type",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemIdentities: success",
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
		expectedStatusCode: http.StatusOK,
		expectedResponse: resources.Response{
			Status: http.StatusOK,
		},
	}, {
		name: "PatchGroupsItemIdentities: failure; invalid JSON",
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: missing request body: request body is not a valid JSON",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PatchGroupsItemIdentities: failure; nil patch array",
		requestBody: resources.GroupIdentitiesPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemIdentities: failure; empty patch array",
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemIdentities: failure; empty identity",
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{{
				Op:       "add",
				Identity: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty identity name",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemIdentities: failure; invalid op",
		requestBody: resources.GroupIdentitiesPatchRequestBody{
			Patches: []resources.GroupIdentitiesPatchItem{{
				Op:       "some-invalid-op",
				Identity: "some-identity",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemIdentities(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: op value not allowed: \"some-invalid-op\"",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemRoles: success",
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
		expectedStatusCode: http.StatusOK,
		expectedResponse: resources.Response{
			Status: http.StatusOK,
		},
	}, {
		name: "PatchGroupsItemRoles: failure; invalid JSON",
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: missing request body: request body is not a valid JSON",
			Status:  http.StatusBadRequest,
		},
	}, {
		name:        "PatchGroupsItemRoles: failure; nil patch array",
		requestBody: resources.GroupRolesPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemRoles: failure; empty patch array",
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty patch array",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemRoles: failure; empty role",
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{{
				Op:   "add",
				Role: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: empty role name",
			Status:  http.StatusBadRequest,
		},
	}, {
		name: "PatchGroupsItemRoles: failure; invalid op",
		requestBody: resources.GroupRolesPatchRequestBody{
			Patches: []resources.GroupRolesPatchItem{{
				Op:   "some-invalid-op",
				Role: "some-role",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchGroupsItemRoles(w, r, "some-id")
		},
		expectedStatusCode: http.StatusBadRequest,
		expectedResponse: resources.Response{
			Message: "Bad Request: invalid request body: op value not allowed: \"some-invalid-op\"",
			Status:  http.StatusBadRequest,
		},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
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
			c.Assert(response.StatusCode, qt.Equals, tt.expectedStatusCode)

			defer response.Body.Close()
			responseBody, err := io.ReadAll(response.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(string(responseBody), qt.JSONEquals, tt.expectedResponse)
		})
	}
}
