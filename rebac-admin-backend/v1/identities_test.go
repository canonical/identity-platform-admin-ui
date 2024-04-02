// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

var (
	mockFirstName  = "MockFirstName"
	mockIdentityId = "test-id"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_identities.go -source=./interfaces/identities.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Identities_ServiceBackendFailures(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErrorResponse := resources.Response{
		Message: "mock-error",
		Status:  http.StatusInternalServerError,
	}

	mockError := errors.New("test-error")

	mockIdentity := resources.Identity{
		Email:   "foo@bar.com",
		Source:  "some-source",
		AddedBy: "some-added-by",
	}

	mockIdentityWithId := resources.Identity{
		Id:      stringPtr("some-id"),
		Email:   "foo@bar.com",
		Source:  "some-source",
		AddedBy: "some-added-by",
	}

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockIdentitiesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestGetIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().ListIdentities(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockParams := resources.GetIdentitiesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/identities", nil)
				h.GetIdentities(w, mockRequest, mockParams)
			},
		},
		{
			name: "TestPostIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().CreateIdentity(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := newTestRequest(http.MethodPost, "/identities", &mockIdentity)
				h.PostIdentities(w, request)
			},
		},
		{
			name: "TestDeleteIdentitiesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().DeleteIdentity(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/identities/%s", mockIdentityId), nil)
				h.DeleteIdentitiesItem(w, mockRequest, "test-id")
			},
		},
		{
			name: "TestGetIdentitiesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentity(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s", mockIdentityId), nil)
				h.GetIdentitiesItem(w, mockRequest, "test-id")
			},
		},
		{
			name: "TestPutIdentitiesItemFailureUpdate",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().UpdateIdentity(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := newTestRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), &mockIdentityWithId)
				h.PutIdentitiesItem(w, request, "test-id")
			},
		},
		{
			name: "TestGetIdentitiesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemEntitlementsParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), nil)
				h.GetIdentitiesItemEntitlements(w, mockRequest, "test-id", params)
			},
		},
		{
			name: "TestPatchIdentitiesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.IdentityEntitlementsPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), patches)
				h.PatchIdentitiesItemEntitlements(w, request, "test-id")
			},
		},
		{
			name: "TestGetIdentitiesItemGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityGroups(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemGroupsParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/groups", mockIdentityId), nil)
				h.GetIdentitiesItemGroups(w, mockRequest, "test-id", params)
			},
		},
		{
			name: "TestPatchIdentitiesItemGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityGroups(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.IdentityGroupsPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/groups", mockIdentityId), patches)
				h.PatchIdentitiesItemGroups(w, request, "test-id")
			},
		},
		{
			name: "TestGetIdentitiesItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemRolesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/roles", mockIdentityId), nil)
				h.GetIdentitiesItemRoles(w, mockRequest, "test-id", params)
			},
		},
		{
			name: "TestPatchIdentitiesItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.IdentityRolesPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, "/identities/test-id/roles", patches)
				h.PatchIdentitiesItemRoles(w, request, "test-id")
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
			tt.setupServiceMock(mockIdentityService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				Identities:            mockIdentityService,
				IdentitiesErrorMapper: mockErrorResponseMapper,
			}

			tt.triggerFunc(sut, mockWriter)

			result := mockWriter.Result()
			defer result.Body.Close()

			c.Assert(result.StatusCode, qt.Equals, http.StatusInternalServerError)

			data, err := io.ReadAll(result.Body)
			c.Assert(err, qt.IsNil)

			response := new(resources.Response)
			err = json.Unmarshal(data, response)

			c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling response, got: %v", err))
			c.Assert(response.Status, qt.Equals, 500)
			c.Assert(response.Message, qt.Equals, "mock-error")
		})
	}
}

func TestHandler_Identities_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityObject := resources.Identity{
		Email:   "foo@bar.com",
		Source:  "some-source",
		AddedBy: "some-added-by",
	}

	mockEntitlements := resources.PaginatedResponse[resources.EntityEntitlement]{
		Data: []resources.EntityEntitlement{
			{
				EntitlementType: "mock-entl-type",
				EntityName:      "mock-entity-name",
				EntityType:      "mock-entity-type",
			},
		},
	}

	mockGroups := resources.PaginatedResponse[resources.Group]{
		Data: []resources.Group{{
			Id:   stringPtr("some-identity-group-id"),
			Name: "some-identity-group-name",
		}},
	}

	mockRoles := resources.PaginatedResponse[resources.Role]{
		Data: []resources.Role{{
			Id:   &mockGroupRoleId,
			Name: mockGroupRoleName,
		}},
	}

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockIdentitiesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Identities_ListIdentitiesSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					ListIdentities(gomock.Any(), gomock.Eq(&resources.GetIdentitiesParams{})).
					Return(&resources.PaginatedResponse[resources.Identity]{
						Data: []resources.Identity{mockIdentityObject},
					}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/identities", nil)
				h.GetIdentities(w, mockRequest, resources.GetIdentitiesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetIdentitiesResponse{
				Data:   []resources.Identity{mockIdentityObject},
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Identities_CreateIdentitySuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					CreateIdentity(gomock.Any(), gomock.Eq(&mockIdentityObject)).
					Return(&mockIdentityObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPost, "/identities", &mockIdentityObject)
				h.PostIdentities(w, mockRequest)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   mockIdentityObject,
		},
		{
			name: "TestHandler_Identities_GetIdentitySuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					GetIdentity(gomock.Any(), gomock.Eq(mockIdentityId)).
					Return(&mockIdentityObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s", mockIdentityId), nil)
				h.GetIdentitiesItem(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockIdentityObject,
		},
		{
			name: "TestHandler_Identities_UpdateIdentitySuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					UpdateIdentity(gomock.Any(), gomock.Eq(&mockIdentityObject)).
					Return(&mockIdentityObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), &mockIdentityObject)
				h.PutIdentitiesItem(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockIdentityObject,
		},
		{
			name: "TestHandler_Identities_DeleteIdentitySuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					DeleteIdentity(gomock.Any(), gomock.Eq(mockIdentityId)).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/identities/%s", mockIdentityId), nil)
				h.DeleteIdentitiesItem(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Identities_GetIdentityGroupsSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					GetIdentityGroups(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&resources.GetIdentitiesItemGroupsParams{})).
					Return(&mockGroups, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/groups", mockIdentityId), nil)
				h.GetIdentitiesItemGroups(w, mockRequest, mockIdentityId, resources.GetIdentitiesItemGroupsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetIdentityGroupsResponse{
				Data:   mockGroups.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Identities_PatchIdentityGroupsSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					PatchIdentityGroups(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.IdentityGroupsPatchRequestBody{
					Patches: []resources.IdentityGroupsPatchItem{
						{
							Group: *mockGroups.Data[0].Id,
							Op:    "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/groups", mockIdentityId), &patches)
				h.PatchIdentitiesItemGroups(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Identities_GetIdentityRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					GetIdentityRoles(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&resources.GetIdentitiesItemRolesParams{})).
					Return(&mockRoles, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/roles", mockIdentityId), nil)
				h.GetIdentitiesItemRoles(w, mockRequest, mockGroupId, resources.GetIdentitiesItemRolesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetIdentityRolesResponse{
				Data:   mockRoles.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Identities_PatchIdentityRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					PatchIdentityRoles(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.IdentityRolesPatchRequestBody{
					Patches: []resources.IdentityRolesPatchItem{
						{
							Role: *mockRoles.Data[0].Id,
							Op:   "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/roles", mockIdentityId), &patches)
				h.PatchIdentitiesItemRoles(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Identities_GetIdentityEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					GetIdentityEntitlements(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&resources.GetIdentitiesItemEntitlementsParams{})).
					Return(&mockEntitlements, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), nil)
				h.GetIdentitiesItemEntitlements(w, mockRequest, mockGroupId, resources.GetIdentitiesItemEntitlementsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetIdentityEntitlementsResponse{
				Data:   mockEntitlements.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Identities_PatchIdentityEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().
					PatchIdentityEntitlements(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.IdentityEntitlementsPatchRequestBody{
					Patches: []resources.IdentityEntitlementsPatchItem{
						{
							Entitlement: mockEntitlements.Data[0],
							Op:          "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), &patches)
				h.PatchIdentitiesItemEntitlements(w, mockRequest, mockIdentityId)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockIdentitiesService := interfaces.NewMockIdentitiesService(ctrl)
			tt.setupServiceMock(mockIdentitiesService)

			sut := handler{Identities: mockIdentitiesService}

			mockWriter := httptest.NewRecorder()
			tt.triggerFunc(sut, mockWriter)

			result := mockWriter.Result()
			defer result.Body.Close()

			c.Assert(result.StatusCode, qt.Equals, tt.expectedStatus)

			body, err := io.ReadAll(result.Body)
			c.Assert(err, qt.IsNil)

			c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling response, got: %v", err))

			if tt.expectedBody != nil {
				c.Assert(string(body), qt.JSONEquals, tt.expectedBody)
			} else {
				c.Assert(len(body), qt.Equals, 0)
			}
		})
	}
}
