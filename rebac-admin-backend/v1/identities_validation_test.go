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

func TestHandlerWithValidation_Identities(t *testing.T) {
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

	validIdentity := resources.Identity{
		Email:   "foo@bar.com",
		Source:  "some-source",
		AddedBy: "some-added-by",
	}

	validIdentityWithId := resources.Identity{
		Id:      stringPtr("some-id"),
		Email:   "foo@bar.com",
		Source:  "some-source",
		AddedBy: "some-added-by",
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
		name:        "PostIdentities: success",
		kind:        kindSuccessful,
		requestBody: validIdentity,
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PostIdentities(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, _ *http.Request) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostIdentities(w, r)
		},
	}, {
		name: "PostIdentities: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostIdentities(w, r)
		},
	}, {
		name: "PostIdentities: failure; empty",
		expectedPatterns: []string{
			"'Email' failed on the 'required' tag",
			"'Source' failed on the 'required' tag",
			"'AddedBy' failed on the 'required' tag",
		},
		requestBody: resources.Identity{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PostIdentities(w, r)
		},
	}, {
		name:        "PutIdentitiesItem: success",
		kind:        kindSuccessful,
		requestBody: validIdentityWithId,
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PutIdentitiesItem(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentitiesItem(w, r, "some-id")
		},
	}, {
		name: "PutIdentitiesItem: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentitiesItem(w, r, "some-id")
		},
	}, {
		name: "PutIdentitiesItem: failure; empty",
		expectedPatterns: []string{
			"'Email' failed on the 'required' tag",
			"'Source' failed on the 'required' tag",
			"'AddedBy' failed on the 'required' tag",
		},
		requestBody: resources.Group{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentitiesItem(w, r, "some-id")
		},
	}, {
		name:             "PutIdentitiesItem: failure; nil id",
		expectedPatterns: []string{"identity ID from path does not match the Identity object"},
		requestBody:      validIdentity,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentitiesItem(w, r, "some-id")
		},
	}, {
		name:             "PutIdentitiesItem: failure; id mismatch",
		expectedPatterns: []string{"identity ID from path does not match the Identity object"},
		requestBody:      validIdentityWithId,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PutIdentitiesItem(w, r, "some-other-id")
		},
	}, {
		name: "PatchIdentitiesItemEntitlements: success",
		kind: kindSuccessful,
		requestBody: resources.IdentityEntitlementsPatchRequestBody{
			Patches: []resources.IdentityEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: validEntitlement,
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchIdentitiesItemEntitlements(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemEntitlements: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemEntitlements: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.IdentityEntitlementsPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemEntitlements: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.IdentityEntitlementsPatchRequestBody{
			Patches: []resources.IdentityEntitlementsPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemEntitlements: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.IdentityEntitlementsPatchRequestBody{
			Patches: []resources.IdentityEntitlementsPatchItem{{
				Op:          "some-invalid-op",
				Entitlement: validEntitlement,
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemEntitlements: failure; invalid entitlement",
		expectedPatterns: []string{
			"'EntitlementType' failed on the 'required' tag",
			"'EntityName' failed on the 'required' tag",
			"'EntityType' failed on the 'required' tag",
		},
		requestBody: resources.IdentityEntitlementsPatchRequestBody{
			Patches: []resources.IdentityEntitlementsPatchItem{{
				Op:          "add",
				Entitlement: resources.EntityEntitlement{},
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemEntitlements(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemGroups: success",
		kind: kindSuccessful,
		requestBody: resources.IdentityGroupsPatchRequestBody{
			Patches: []resources.IdentityGroupsPatchItem{{
				Op:    "add",
				Group: "some-group",
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchIdentitiesItemGroups(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemGroups: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemGroups: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.IdentityGroupsPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemGroups: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.IdentityGroupsPatchRequestBody{
			Patches: []resources.IdentityGroupsPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemGroups: failure; empty identity",
		expectedPatterns: []string{"'Group' failed on the 'required' tag"},
		requestBody: resources.IdentityGroupsPatchRequestBody{
			Patches: []resources.IdentityGroupsPatchItem{{
				Op:    "add",
				Group: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemGroups: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.IdentityGroupsPatchRequestBody{
			Patches: []resources.IdentityGroupsPatchItem{{
				Op:    "some-invalid-op",
				Group: "some-group",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemGroups(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemRoles: success",
		kind: kindSuccessful,
		requestBody: resources.IdentityRolesPatchRequestBody{
			Patches: []resources.IdentityRolesPatchItem{{
				Op:   "add",
				Role: "some-role",
			}},
		},
		setupHandlerMock: func(mockHandler *resources.MockServerInterface) {
			mockHandler.EXPECT().
				PatchIdentitiesItemRoles(gomock.Any(), gomock.Any(), "some-id").
				Do(func(w http.ResponseWriter, _ *http.Request, _ string) {
					writeResponse(w, http.StatusOK, resources.Response{
						Status: http.StatusOK,
					})
				})
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
		},
	}, {
		name: "PatchIdentitiesItemRoles: failure; invalid JSON",
		kind: kindBadJSON,
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemRoles: failure; nil patch array",
		expectedPatterns: []string{"'Patches' failed on the 'required' tag"},
		requestBody:      resources.IdentityRolesPatchRequestBody{},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemRoles: failure; empty patch array",
		expectedPatterns: []string{"'Patches' failed on the 'gt' tag"},
		requestBody: resources.IdentityRolesPatchRequestBody{
			Patches: []resources.IdentityRolesPatchItem{},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemRoles: failure; empty role",
		expectedPatterns: []string{"'Role' failed on the 'required' tag"},
		requestBody: resources.IdentityRolesPatchRequestBody{
			Patches: []resources.IdentityRolesPatchItem{{
				Op:   "add",
				Role: "",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
		},
	}, {
		name:             "PatchIdentitiesItemRoles: failure; invalid op",
		expectedPatterns: []string{"'Op' failed on the 'oneof' tag"},
		requestBody: resources.IdentityRolesPatchRequestBody{
			Patches: []resources.IdentityRolesPatchItem{{
				Op:   "some-invalid-op",
				Role: "some-role",
			}},
		},
		triggerFunc: func(sut *handlerWithValidation, w http.ResponseWriter, r *http.Request) {
			sut.PatchIdentitiesItemRoles(w, r, "some-id")
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
