// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
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

	mockEntitlements := []resources.EntityEntitlement{
		{
			EntitlementType: "mock-entl-type",
			EntityName:      "mock-entity-name",
			EntityType:      "mock-entity-type",
		},
	}

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockRolesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		doAssert         func(c *qt.C, result *http.Response)
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Roles_ListRolesSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					ListRoles(gomock.Any(), gomock.Eq(&resources.GetRolesParams{})).
					Return(&resources.Roles{
						Data: []resources.Role{mockRoleObject},
					}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/roles", nil)
				h.GetRoles(w, mockRequest, resources.GetRolesParams{})
			},
			expectedStatus: http.StatusOK,
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				response := new(resources.GetRolesResponse)
				err = json.Unmarshal(data, response)

				c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

				expected := resources.GetRolesResponse{
					Data:   []resources.Role{mockRoleObject},
					Status: http.StatusOK,
				}
				c.Assert(*response, qt.DeepEquals, expected)
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
				roleBody, _ := json.Marshal(mockRoleObject)
				mockRequest := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
				h.PostRoles(w, mockRequest)
			},
			expectedStatus: http.StatusCreated,
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				response := new(resources.Role)
				err = json.Unmarshal(data, response)

				c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

				expected := mockRoleObject
				c.Assert(*response, qt.DeepEquals, expected)
			},
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
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				response := new(resources.Role)
				err = json.Unmarshal(data, response)

				c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

				expected := mockRoleObject
				c.Assert(*response, qt.DeepEquals, expected)
			},
		},
		{
			name: "TestHandler_Roles_UpdateRoleSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					UpdateRole(gomock.Any(), gomock.Eq(&mockRoleObject)).
					Return(&mockRoleObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				roleBody, _ := json.Marshal(mockRoleObject)
				mockRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/roles/%s", mockRoleId), bytes.NewReader(roleBody))
				h.PutRolesItem(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				response := new(resources.Role)
				err = json.Unmarshal(data, response)

				c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

				expected := mockRoleObject
				c.Assert(*response, qt.DeepEquals, expected)
			},
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
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
				c.Assert(len(data), qt.Equals, 0)
			},
		},
		{
			name: "TestHandler_Roles_GetRoleEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().
					GetRoleEntitlements(gomock.Any(), gomock.Eq(mockRoleId), gomock.Eq(&resources.GetRolesItemEntitlementsParams{})).
					Return(mockEntitlements, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), nil)
				h.GetRolesItemEntitlements(w, mockRequest, mockRoleId, resources.GetRolesItemEntitlementsParams{})
			},
			expectedStatus: http.StatusOK,
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				response := new(resources.GetRoleEntitlementsResponse)
				err = json.Unmarshal(data, response)

				c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

				expected := resources.GetRoleEntitlementsResponse{
					Data:   mockEntitlements,
					Status: http.StatusOK,
				}
				c.Assert(*response, qt.DeepEquals, expected)
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
				patchesBody, _ := json.Marshal(resources.RoleEntitlementsPatchRequestBody{
					Patches: []resources.RoleEntitlementsPatchItem{
						{
							Entitlement: mockEntitlements[0],
							Op:          "add",
						},
					},
				})
				mockRequest := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), bytes.NewReader(patchesBody))
				h.PatchRolesItemEntitlements(w, mockRequest, mockRoleId)
			},
			expectedStatus: http.StatusOK,
			doAssert: func(c *qt.C, result *http.Response) {
				data, err := io.ReadAll(result.Body)
				c.Assert(err, qt.IsNil)

				c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
				c.Assert(len(data), qt.Equals, 0)
			},
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
			tt.doAssert(c, result)
		})
	}

}

func TestHandler_Roles_ValidationErrors(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// need a value that is not a struct to trigger Decode error
	mockInvalidRequestBody := true

	invalidRequestBody, _ := json.Marshal(mockInvalidRequestBody)

	type EndpointTest struct {
		name        string
		triggerFunc func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestPostRolesFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/roles/%s", mockRoleId), bytes.NewReader(invalidRequestBody))
				h.PostRoles(w, req)
			},
		},
		{
			name: "TestPutRolesFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/roles/%s", mockRoleId), bytes.NewReader(invalidRequestBody))
				h.PutRolesItem(w, req, mockRoleId)
			},
		},
		{
			name: "TestPatchRolesEntitlementsFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), bytes.NewReader(invalidRequestBody))
				h.PatchRolesItemEntitlements(w, req, mockRoleId)
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockWriter := httptest.NewRecorder()
			sut := handler{}

			tt.triggerFunc(sut, mockWriter)

			result := mockWriter.Result()
			defer result.Body.Close()

			c.Assert(result.StatusCode, qt.Equals, http.StatusBadRequest)

			data, err := io.ReadAll(result.Body)
			c.Assert(err, qt.IsNil)

			response := new(resources.Response)

			err = json.Unmarshal(data, response)
			c.Assert(err, qt.IsNil)

			c.Assert(response.Status, qt.Equals, http.StatusBadRequest)
			c.Assert(response.Message, qt.Equals, "Bad Request: request doesn't match the expected schema")
		})
	}
}

// OK
func TestHandler_Roles_ServiceBackendFailures(t *testing.T) {
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
		setupServiceMock func(mockService *interfaces.MockRolesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		skip             bool
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
				role, _ := json.Marshal(&resources.Role{})
				request := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(role))
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
				h.DeleteRolesItem(w, mockRequest, "test-id")
			},
		},
		{
			name: "TestGetRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().GetRole(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s", mockRoleId), nil)
				h.GetRolesItem(w, mockRequest, "test-id")
			},
		},
		{
			name: "TestPutRolesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().UpdateRole(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				role, _ := json.Marshal(&resources.Role{Id: &mockRoleId})
				request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/roles/%s", mockRoleId), bytes.NewReader(role))
				h.PutRolesItem(w, request, "test-id")
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
				h.GetRolesItemEntitlements(w, mockRequest, "test-id", params)
			},
		},
		{
			name: "TestPatchIdentitiesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockRolesService) {
				mockService.EXPECT().PatchRoleEntitlements(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches, _ := json.Marshal(&resources.RoleEntitlementsPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/roles/%s/entitlements", mockRoleId), bytes.NewReader(patches))
				h.PatchRolesItemEntitlements(w, request, "test-id")
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

			c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
			c.Assert(response.Status, qt.Equals, http.StatusInternalServerError)
			c.Assert(response.Message, qt.Equals, "mock-error")
		})
	}
}
