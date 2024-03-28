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

	"go.uber.org/mock/gomock"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

var (
	mockRoleName = "MockRoleName"
	mockRoleId   = "test-id"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_roles.go -source=./interfaces/roles.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Roles_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleObject := resources.Role{
		Id:   &mockRoleId,
		Name: mockRoleName,
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

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockRolesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Roles_ListRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					ListRoles(gomock.Any(), gomock.Eq(&resources.GetRolesParams{})).
					Return(&resources.PaginatedResponse[resources.Role]{
						Data: []resources.Role{mockRoleObject},
					}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/roles", nil)
				h.GetRoles(w, mockRequest, resources.GetRolesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetRolesResponse{
				Data:   []resources.Role{mockRoleObject},
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Roles_CreateRoleSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					CreateRole(gomock.Any(), gomock.Eq(&mockRoleObject)).
					Return(&mockRoleObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPost, "/roles", &mockRoleObject)
				h.PostRoles(w, mockRequest)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   mockRoleObject,
		},
		{
			name: "TestHandler_Roles_GetRoleSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					GetRole(gomock.Any(), gomock.Eq(mockRoleId)).
					Return(&mockRoleObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s", mockRoleId), nil)
				h.GetRolesItem(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockRoleObject,
		},
		{
			name: "TestHandler_Roles_UpdateRoleSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					UpdateRole(gomock.Any(), gomock.Eq(&mockRoleObject)).
					Return(&mockRoleObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := newTestRequest(http.MethodPut, fmt.Sprintf("/roles/%s", mockRoleId), &mockRoleObject)
				h.PutRolesItem(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockRoleObject,
		},
		{
			name: "TestHandler_Roles_DeleteRoleSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					DeleteRole(gomock.Any(), gomock.Eq(mockRoleId)).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/roles/%s", mockRoleId), nil)
				h.DeleteRolesItem(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_Roles_GetRoleEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					GetRoleEntitlements(gomock.Any(), gomock.Eq(mockRoleId), gomock.Eq(&resources.GetRolesItemEntitlementsParams{})).
					Return(&mockEntitlements, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), nil)
				h.GetRolesItemEntitlements(w, mockRequest, mockRoleId, resources.GetRolesItemEntitlementsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetRoleEntitlementsResponse{
				Data:   mockEntitlements.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Roles_PatchRoleEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					PatchRoleEntitlements(gomock.Any(), gomock.Eq(mockRoleId), gomock.Any()).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := resources.RoleEntitlementsPatchRequestBody{
					Patches: []resources.RoleEntitlementsPatchItem{
						{
							Entitlement: mockEntitlements.Data[0],
							Op:          "add",
						},
					},
				}
				mockRequest := newTestRequest(http.MethodPatch, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), &patches)
				h.PatchRolesItemEntitlements(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockRoleService := interfaces.NewMockRolesService(ctrl)
			tt.setupServiceMock(mockRoleService)

			sut := handler{Roles: mockRoleService}

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

func TestHandler_Roles_ServiceBackendFailures(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErrorResponse := resources.Response{
		Message: "mock-error",
		Status:  http.StatusInternalServerError,
	}

	mockError := errors.New("test-error")

	mockRoleObject := resources.Role{
		Id:   &mockRoleId,
		Name: mockRoleName,
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

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockRolesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestGetRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockParams := resources.GetRolesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/roles", nil)
				h.GetRoles(w, mockRequest, mockParams)
			},
		},
		{
			name: "TestPostRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().CreateRole(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := newTestRequest(http.MethodPost, "/roles", &mockRoleObject)
				h.PostRoles(w, request)
			},
		},
		{
			name: "TestDeleteRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().DeleteRole(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/roles/%s", mockRoleId), nil)
				h.DeleteRolesItem(w, mockRequest, mockRoleId)
			},
		},
		{
			name: "TestGetRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().GetRole(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s", mockRoleId), nil)
				h.GetRolesItem(w, mockRequest, mockRoleId)
			},
		},
		{
			name: "TestPutRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().UpdateRole(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := newTestRequest(http.MethodPut, fmt.Sprintf("/roles/%s", mockRoleId), &mockRoleObject)
				h.PutRolesItem(w, request, mockRoleId)
			},
		},
		{
			name: "TestGetRolesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().GetRoleEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetRolesItemEntitlementsParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), nil)
				h.GetRolesItemEntitlements(w, mockRequest, mockRoleId, params)
			},
		},
		{
			name: "TestPatchRolesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().PatchRoleEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches := &resources.RoleEntitlementsPatchRequestBody{
					Patches: []resources.RoleEntitlementsPatchItem{{
						Op:          "add",
						Entitlement: mockEntitlements.Data[0],
					}},
				}
				request := newTestRequest(http.MethodPatch, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), patches)
				h.PatchRolesItemEntitlements(w, request, mockRoleId)
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockRoleService := interfaces.NewMockRolesService(ctrl)
			tt.setupServiceMock(mockRoleService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				Roles:            mockRoleService,
				RolesErrorMapper: mockErrorResponseMapper,
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
