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
	mockGroupName              = "MockGroupName"
	mockGroupIdentityFirstName = "MockGroupIdentityFirstName"
	mockGroupIdentityId        = "MockGroupIdentityId"
	mockGroupRoleId            = "MockRoleId"
	mockGroupRoleName          = "MockRoleName"
	mockGroupId                = "test-id"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_groups.go -source=./interfaces/groups.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Groups_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupObject := resources.Group{
		Id:   &mockGroupId,
		Name: mockGroupName,
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

	mockIdentities := resources.PaginatedResponse[resources.Identity]{
		Data: []resources.Identity{{
			Id:        &mockGroupIdentityId,
			FirstName: &mockGroupIdentityFirstName,
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
		setupServiceMock func(mockService *interfaces.MockGroupsService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Groups_ListGroupsSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					ListGroups(gomock.Any(), gomock.Eq(&resources.GetGroupsParams{})).
					Return(&resources.PaginatedResponse[resources.Group]{
						Data: []resources.Group{mockGroupObject},
					}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/groups", nil)
				h.GetGroups(w, mockRequest, resources.GetGroupsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetGroupsResponse{
				Data:   []resources.Group{mockGroupObject},
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Groups_CreateGroupSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					CreateGroup(gomock.Any(), gomock.Eq(&mockGroupObject)).
					Return(&mockGroupObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPost, "/groups", &mockGroupObject)
				h.PostGroups(w, mockRequest)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   mockGroupObject,
		},
		{
			name: "TestHandler_Groups_GetGroupSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					GetGroup(gomock.Any(), gomock.Eq(mockGroupId)).
					Return(&mockGroupObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s", mockGroupId), nil)
				h.GetGroupsItem(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockGroupObject,
		},
		{
			name: "TestHandler_Groups_UpdateGroupSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					UpdateGroup(gomock.Any(), gomock.Eq(&mockGroupObject)).
					Return(&mockGroupObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPut, fmt.Sprintf("/groups/%s", mockGroupId), &mockGroupObject)
				h.PutGroupsItem(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockGroupObject,
		},
		{
			name: "TestHandler_Groups_DeleteGroupSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					DeleteGroup(gomock.Any(), gomock.Eq(mockGroupId)).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/%s", mockGroupId), nil)
				h.DeleteGroupsItem(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Groups_GetGroupIdentitiesSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					GetGroupIdentities(gomock.Any(), gomock.Eq(mockGroupId), gomock.Eq(&resources.GetGroupsItemIdentitiesParams{})).
					Return(&mockIdentities, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/identities", mockGroupId), nil)
				h.GetGroupsItemIdentities(w, mockRequest, mockGroupId, resources.GetGroupsItemIdentitiesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetGroupIdentitiesResponse{
				Data:   mockIdentities.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Groups_PatchGroupIdentitiesSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					PatchGroupIdentities(gomock.Any(), gomock.Eq(mockGroupId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.GroupIdentitiesPatchRequestBody{
					Patches: []resources.GroupIdentitiesPatchItem{
						{
							Identity: *mockIdentities.Data[0].Id,
							Op:       "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/identities", mockGroupId), &patches)
				h.PatchGroupsItemIdentities(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Groups_GetGroupRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					GetGroupRoles(gomock.Any(), gomock.Eq(mockGroupId), gomock.Eq(&resources.GetGroupsItemRolesParams{})).
					Return(&mockRoles, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/roles", mockGroupId), nil)
				h.GetGroupsItemRoles(w, mockRequest, mockGroupId, resources.GetGroupsItemRolesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetGroupRolesResponse{
				Data:   mockRoles.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Groups_PatchGroupRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					PatchGroupRoles(gomock.Any(), gomock.Eq(mockGroupId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.GroupRolesPatchRequestBody{
					Patches: []resources.GroupRolesPatchItem{
						{
							Role: *mockRoles.Data[0].Id,
							Op:   "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/roles", mockGroupId), &patches)
				h.PatchGroupsItemRoles(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Groups_GetGroupEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					GetGroupEntitlements(gomock.Any(), gomock.Eq(mockGroupId), gomock.Eq(&resources.GetGroupsItemEntitlementsParams{})).
					Return(&mockEntitlements, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/entitlements", mockGroupId), nil)
				h.GetGroupsItemEntitlements(w, mockRequest, mockGroupId, resources.GetGroupsItemEntitlementsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetGroupEntitlementsResponse{
				Data:   mockEntitlements.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Groups_PatchGroupEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().
					PatchGroupEntitlements(gomock.Any(), gomock.Eq(mockGroupId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.GroupEntitlementsPatchRequestBody{
					Patches: []resources.GroupEntitlementsPatchItem{
						{
							Entitlement: mockEntitlements.Data[0],
							Op:          "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/entitlements", mockGroupId), &patches)
				h.PatchGroupsItemEntitlements(w, mockRequest, mockGroupId)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockGroupsService := interfaces.NewMockGroupsService(ctrl)
			tt.setupServiceMock(mockGroupsService)

			sut := handler{Groups: mockGroupsService}

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

func TestHandler_Groups_ServiceBackendFailures(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErrorResponse := resources.Response{
		Message: "mock-error",
		Status:  http.StatusInternalServerError,
	}

	mockError := errors.New("test-error")

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockGroupsService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestGetGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().ListGroups(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockParams := resources.GetGroupsParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/groups", nil)
				h.GetGroups(w, mockRequest, mockParams)
			},
		},
		{
			name: "TestPostGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().CreateGroup(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockGroup := &resources.Group{}
				request := newTestRequest(http.MethodPost, "/groups", mockGroup)
				h.PostGroups(w, request)
			},
		},
		{
			name: "TestDeleteGroupsItemFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().DeleteGroup(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/%s", mockGroupId), nil)
				h.DeleteGroupsItem(w, mockRequest, mockGroupId)
			},
		},
		{
			name: "TestGetRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().GetGroup(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s", mockGroupId), nil)
				h.GetGroupsItem(w, mockRequest, mockGroupId)
			},
		},
		{
			name: "TestPutGroupsItemFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().UpdateGroup(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockGroup := &resources.Group{Id: &mockGroupId}
				request := newTestRequest(http.MethodPut, fmt.Sprintf("/groups/%s", mockGroupId), mockGroup)
				h.PutGroupsItem(w, request, mockGroupId)
			},
		},
		{
			name: "TestGetGroupsItemIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().GetGroupIdentities(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetGroupsItemIdentitiesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/identities", mockGroupId), nil)
				h.GetGroupsItemIdentities(w, mockRequest, mockGroupId, params)
			},
		},
		{
			name: "TestPatchGroupsItemIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().PatchGroupIdentities(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.GroupIdentitiesPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/identities", mockGroupId), patches)
				h.PatchGroupsItemIdentities(w, request, mockGroupId)
			},
		},
		{
			name: "TestGetGroupsItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().GetGroupRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetGroupsItemRolesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/roles", mockGroupId), nil)
				h.GetGroupsItemRoles(w, mockRequest, mockGroupId, params)
			},
		},
		{
			name: "TestPatchGroupsItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().PatchGroupRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.GroupRolesPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/roles", mockGroupId), patches)
				h.PatchGroupsItemRoles(w, request, mockGroupId)
			},
		},
		{
			name: "TestGetGroupsItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().GetGroupEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetGroupsItemEntitlementsParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/groups/%s/entitlements", mockGroupId), nil)
				h.GetGroupsItemEntitlements(w, mockRequest, mockGroupId, params)
			},
		},
		{
			name: "TestPatchGroupsItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockGroupsService) {
				mockService.EXPECT().PatchGroupEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.GroupEntitlementsPatchRequestBody{}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/groups/%s/entitlements", mockGroupId), patches)
				h.PatchGroupsItemEntitlements(w, request, mockGroupId)
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockGroupsService := interfaces.NewMockGroupsService(ctrl)
			tt.setupServiceMock(mockGroupsService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				Groups:            mockGroupsService,
				GroupsErrorMapper: mockErrorResponseMapper,
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
			c.Assert(response.Status, qt.Equals, http.StatusInternalServerError)
			c.Assert(response.Message, qt.Equals, "mock-error")
		})
	}
}
