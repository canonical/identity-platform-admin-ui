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
	mockFirstName  = "MockFirstName"
	mockIdentityId = "test-id"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_identities.go -source=./interfaces/identities.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_GetIdentities(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParams := resources.GetIdentitiesParams{}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentitiesReturn := resources.Identities{Data: []resources.Identity{
		{FirstName: &mockFirstName},
	}}
	mockIdentityService.EXPECT().ListIdentities(gomock.Any(), gomock.Any()).Return(&mockIdentitiesReturn, nil)

	expectedResponse := resources.GetIdentitiesResponse{
		Data:   mockIdentitiesReturn.Data,
		Status: 200,
	}

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, "/identities", nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentities(mockWriter, mockRequest, mockParams)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.GetIdentitiesResponse)

	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PostIdentitiesSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityReturn := resources.Identity{FirstName: &mockFirstName}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().CreateIdentity(gomock.Any(), gomock.Any()).Return(&mockIdentityReturn, nil)

	marshalledIdentity, err := json.Marshal(mockIdentityReturn)
	c.Assert(err, qt.IsNil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPost, "/identities", bytes.NewReader(marshalledIdentity))

	sut := handler{Identities: mockIdentityService}
	sut.PostIdentities(mockWriter, mockRequest)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.Identity)
	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusCreated)
	c.Assert(response, qt.DeepEquals, &mockIdentityReturn)
}

func TestHandler_DeleteIdentitiesItemSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().DeleteIdentity(gomock.Any(), gomock.Eq(mockIdentityId)).Return(true, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/identities/%s", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.DeleteIdentitiesItem(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(len(data), qt.Equals, 0)
}

func TestHandler_GetIdentitiesItemSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityReturn := resources.Identity{FirstName: &mockFirstName}
	mockIdentityService.EXPECT().GetIdentity(gomock.Any(), gomock.Eq(mockIdentityId)).Return(&mockIdentityReturn, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItem(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.Identity)

	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &mockIdentityReturn)
}

func TestHandler_PutIdentitiesItemSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityReturn := resources.Identity{
		Id:        &mockIdentityId,
		FirstName: &mockFirstName,
	}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().UpdateIdentity(gomock.Any(), gomock.Any()).Return(&mockIdentityReturn, nil)

	marshalledIdentity, err := json.Marshal(mockIdentityReturn)
	c.Assert(err, qt.IsNil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), bytes.NewReader(marshalledIdentity))

	sut := handler{Identities: mockIdentityService}
	sut.PutIdentitiesItem(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.Identity)
	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &mockIdentityReturn)
}

func TestHandler_GetIdentitiesItemEntitlementsSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		entitlementType = "mock-entl-type"
		entityName      = "mock-entity-name"
		entityType      = "mock-entity-type"
	)

	mockIdentityEntitlements := []resources.EntityEntitlement{
		{
			EntitlementType: entitlementType,
			EntityName:      entityName,
			EntityType:      entityType,
		},
	}
	expectedResponse := resources.GetIdentityEntitlementsResponse{
		Data:   mockIdentityEntitlements,
		Status: http.StatusOK,
	}

	params := resources.GetIdentitiesItemEntitlementsParams{}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().GetIdentityEntitlements(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(mockIdentityEntitlements, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemEntitlements(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.GetIdentityEntitlementsResponse)
	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemEntitlementsSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	identityEntitlementPatches := []resources.IdentityEntitlementsPatchItem{
		{
			Entitlement: resources.EntityEntitlement{},
			Op:          "add",
		},
	}
	identityEntitlementPatchRequest := resources.IdentityEntitlementsPatchRequestBody{
		Patches: identityEntitlementPatches,
	}

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().PatchIdentityEntitlements(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(identityEntitlementPatches)).Return(true, nil)

	marshalledPatchReq, err := json.Marshal(identityEntitlementPatchRequest)
	c.Assert(err, qt.IsNil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), bytes.NewReader(marshalledPatchReq))

	sut := handler{Identities: mockIdentityService}
	sut.PatchIdentitiesItemEntitlements(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
}

func TestHandler_GetIdentitiesItemGroupsSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		mockId   = "mock-id"
		mockName = "test-groupname"
	)

	mockIdentityGroups := resources.Groups{
		Data: []resources.Group{
			{
				Id:   &mockId,
				Name: mockName,
			},
		},
	}
	expectedResponse := resources.GetIdentityGroupsResponse{
		Data:   mockIdentityGroups.Data,
		Status: http.StatusOK,
	}

	params := resources.GetIdentitiesItemGroupsParams{}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().GetIdentityGroups(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(&mockIdentityGroups, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/groups", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemGroups(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.GetIdentityGroupsResponse)
	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemGroupsSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	identityGroupsPatches := []resources.IdentityGroupsPatchItem{
		{
			Group: "test-group-identifier",
			Op:    "add",
		},
	}
	identityGroupsPatchRequest := resources.IdentityGroupsPatchRequestBody{
		Patches: identityGroupsPatches,
	}

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().PatchIdentityGroups(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(identityGroupsPatches)).Return(true, nil)

	marshalledPatchReq, err := json.Marshal(identityGroupsPatchRequest)
	c.Assert(err, qt.IsNil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/groups", mockIdentityId), bytes.NewReader(marshalledPatchReq))

	sut := handler{Identities: mockIdentityService}
	sut.PatchIdentitiesItemGroups(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
}

func TestHandler_GetIdentitiesItemRolesSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		mockId   = "mock-id"
		mockName = "test-rolename"
	)

	mockIdentityRoles := resources.Roles{
		Data: []resources.Role{
			{
				Id:   &mockId,
				Name: mockName,
			},
		},
	}
	expectedResponse := resources.GetIdentityRolesResponse{
		Data:   mockIdentityRoles.Data,
		Status: http.StatusOK,
	}

	params := resources.GetIdentitiesItemRolesParams{}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().GetIdentityRoles(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(&mockIdentityRoles, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/roles", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemRoles(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	response := new(resources.GetIdentityRolesResponse)
	err = json.Unmarshal(data, response)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemRolesSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	identityRolesPatches := []resources.IdentityRolesPatchItem{
		{
			Role: "test-role-identifier",
			Op:   "add",
		},
	}
	identityRolesPatchRequest := resources.IdentityRolesPatchRequestBody{
		Patches: identityRolesPatches,
	}

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().PatchIdentityRoles(gomock.Any(), gomock.Eq(mockIdentityId), gomock.Eq(identityRolesPatches)).Return(true, nil)

	marshalledPatchReq, err := json.Marshal(identityRolesPatchRequest)
	c.Assert(err, qt.IsNil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/roles", mockIdentityId), bytes.NewReader(marshalledPatchReq))

	sut := handler{Identities: mockIdentityService}
	sut.PatchIdentitiesItemRoles(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
}

func TestHandler_Identities_ValidationErrors(t *testing.T) {
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
			name: "TestPatchIdentitiesEntitlementsFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), bytes.NewReader(invalidRequestBody))
				h.PatchIdentitiesItemEntitlements(w, req, mockIdentityId)
			},
		},
		{
			name: "TestPatchIdentitiesGroupsFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/groups", mockIdentityId), bytes.NewReader(invalidRequestBody))
				h.PatchIdentitiesItemGroups(w, req, mockIdentityId)
			},
		},
		{
			name: "TestPatchIdentitiesRolesFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/roles", mockIdentityId), bytes.NewReader(invalidRequestBody))
				h.PatchIdentitiesItemRoles(w, req, mockIdentityId)
			},
		},
		{
			name: "TestPostIdentitiesFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/identities/%s", mockIdentityId), bytes.NewReader(invalidRequestBody))
				h.PostIdentities(w, req)
			},
		},
		{
			name: "TestPutIdentitiesFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), bytes.NewReader(invalidRequestBody))
				h.PutIdentitiesItem(w, req, mockIdentityId)
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

func TestHandler_Identities_Failures(t *testing.T) {
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
				identity, _ := json.Marshal(&resources.Identity{})
				request := httptest.NewRequest(http.MethodPost, "/identities", bytes.NewReader(identity))
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
				identity, _ := json.Marshal(&resources.Identity{Id: &mockIdentityId})
				request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), bytes.NewReader(identity))
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
				patches, _ := json.Marshal(&resources.IdentityEntitlementsPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), bytes.NewReader(patches))
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
				patches, _ := json.Marshal(&resources.IdentityGroupsPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/identities/%s/groups", mockIdentityId), bytes.NewReader(patches))
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
				patches, _ := json.Marshal(&resources.IdentityRolesPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, "/identities/test-id/roles", bytes.NewReader(patches))
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

			c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
			c.Assert(response.Status, qt.Equals, 500)
			c.Assert(response.Message, qt.Equals, "mock-error")
		})
	}
}
