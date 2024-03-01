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
	mockFirstName  string = "MockFirstName"
	mockIdentityId string = "test-id"
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
	mockIdentityService.EXPECT().ListIdentities(gomock.Any()).Return(&mockIdentitiesReturn, nil)

	expectedResponse := resources.GetIdentitiesResponse{
		Data:   mockIdentitiesReturn.Data,
		Status: 200,
	}

	mockWriter := httptest.NewRecorder()

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentities(mockWriter, nil, mockParams)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.GetIdentitiesResponse)

	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling GetIdentities response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PostIdentitiesSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityReturn := resources.Identity{FirstName: &mockFirstName}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().CreateIdentity(gomock.Any()).Return(&mockIdentityReturn, nil)

	marshalledIdentity, err := json.Marshal(mockIdentityReturn)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPost, "/identities", bytes.NewReader(marshalledIdentity))

	sut := handler{Identities: mockIdentityService}
	sut.PostIdentities(mockWriter, mockRequest)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.Identity)
	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling PostIdentities response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusCreated)
	c.Assert(response, qt.DeepEquals, &mockIdentityReturn)
}

func TestHandler_DeleteIdentitiesItemSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().DeleteIdentity(gomock.Eq(mockIdentityId)).Return(true, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/identities/%s", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.DeleteIdentitiesItem(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(len(data), qt.Equals, 0)
}

func TestHandler_GetIdentitiesItemSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityReturn := resources.Identity{FirstName: &mockFirstName}
	mockIdentityService.EXPECT().GetIdentity(gomock.Eq(mockIdentityId)).Return(&mockIdentityReturn, nil)

	mockWriter := httptest.NewRecorder()

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItem(mockWriter, nil, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.Identity)

	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling GetIdentitiesItem response, got: %v", err)
	}

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
	mockIdentityService.EXPECT().UpdateIdentity(gomock.Any()).Return(&mockIdentityReturn, nil)

	marshalledIdentity, err := json.Marshal(mockIdentityReturn)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/identities/%s", mockIdentityId), bytes.NewReader(marshalledIdentity))

	sut := handler{Identities: mockIdentityService}
	sut.PutIdentitiesItem(mockWriter, mockRequest, mockIdentityId)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.Identity)
	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling PutIdentitiesItem response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &mockIdentityReturn)
}

func TestHandler_GetIdentitiesItemEntitlementsSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		mockId          = "mock-id"
		mockEntity      = resources.Entity{"entity-name": "mock-entity-name"}
		mockEntitlement = "test-entitlement"
	)

	mockIdentityEntitlements := resources.EntityEntitlements{
		Data: []resources.EntityEntitlement{
			{
				Id:          &mockId,
				Entity:      &mockEntity,
				Entitlement: &mockEntitlement,
			},
		},
	}
	expectedResponse := resources.GetIdentityEntitlementsResponse{
		Data:   mockIdentityEntitlements.Data,
		Status: http.StatusOK,
	}

	params := resources.GetIdentitiesItemEntitlementsParams{}
	mockIdentityService := interfaces.NewMockIdentitiesService(ctrl)
	mockIdentityService.EXPECT().GetIdentityEntitlements(gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(&mockIdentityEntitlements, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/entitlements", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemEntitlements(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.GetIdentityEntitlementsResponse)
	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling GetIdentitiesItemEntitlements response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemEntitlementsSuccess(t *testing.T) {
	//
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
	mockIdentityService.EXPECT().GetIdentityGroups(gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(&mockIdentityGroups, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/groups", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemGroups(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.GetIdentityGroupsResponse)
	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling GetIdentitiesItemGroups response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemGroupsSuccess(t *testing.T) {

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
	mockIdentityService.EXPECT().GetIdentityRoles(gomock.Eq(mockIdentityId), gomock.Eq(&params)).Return(&mockIdentityRoles, nil)

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/identities/%s/roles", mockIdentityId), nil)

	sut := handler{Identities: mockIdentityService}
	sut.GetIdentitiesItemRoles(mockWriter, mockRequest, mockIdentityId, params)

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.GetIdentityRolesResponse)
	if err := json.Unmarshal(data, response); err != nil {
		t.Errorf("Unexpected err while unmarshaling GetIdentitiesItemRoles response, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(response, qt.DeepEquals, &expectedResponse)
}

func TestHandler_PatchIdentitiesItemRolesSuccess(t *testing.T) {

}

func TestHandler_Failures(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParams := resources.GetIdentitiesParams{}
	mockErrorResponse := resources.Response{
		Message: "mock-error",
		Status:  http.StatusInternalServerError,
	}

	mockError := errors.New("test-error")

	for _, test := range []struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockIdentitiesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		skip             bool
	}{
		{
			name: "TestGetIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().ListIdentities(gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				h.GetIdentities(w, nil, mockParams)
			},
			skip: false,
		},
		{
			name: "TestPostIdentitiesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().CreateIdentity(gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				identity, _ := json.Marshal(&resources.Identity{})
				request := httptest.NewRequest(http.MethodPost, "/identities", bytes.NewReader(identity))
				h.PostIdentities(w, request)
			},
			skip: false,
		},
		{
			name: "TestDeleteIdentitiesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().DeleteIdentity(gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				h.DeleteIdentitiesItem(w, nil, "test-id")
			},
			skip: false,
		},
		{
			name: "TestGetIdentitiesItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentity(gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				h.GetIdentitiesItem(w, nil, "test-id")
			},
			skip: false,
		},
		{
			name: "TestPutIdentitiesItemFailureUpdate",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().UpdateIdentity(gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				identity, _ := json.Marshal(&resources.Identity{Id: &mockIdentityId})
				request := httptest.NewRequest(http.MethodPut, "/identities", bytes.NewReader(identity))
				h.PutIdentitiesItem(w, request, "test-id")
			},
			skip: false,
		},
		{
			name: "TestGetIdentitiesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityEntitlements(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemEntitlementsParams{}
				h.GetIdentitiesItemEntitlements(w, nil, "test-id", params)
			},
			skip: false,
		},
		{
			name: "TestPatchIdentitiesItemEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityEntitlements(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches, _ := json.Marshal(&resources.EntityEntitlementPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, "/identities/test-id/entitlements", bytes.NewReader(patches))
				h.PatchIdentitiesItemEntitlements(w, request, "test-id")
			},
			skip: true,
		},
		{
			name: "TestGetIdentitiesItemGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityGroups(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemGroupsParams{}
				h.GetIdentitiesItemGroups(w, nil, "test-id", params)
			},
			skip: false,
		},
		{
			name: "TestPatchIdentitiesItemGroupsFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityGroups(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches, _ := json.Marshal(&resources.EntityEntitlementPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, "/identities/test-id/groups", bytes.NewReader(patches))
				h.PatchIdentitiesItemGroups(w, request, "test-id")
			},
			skip: true,
		},
		{
			name: "TestGetIdentitiesItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().GetIdentityRoles(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentitiesItemRolesParams{}
				h.GetIdentitiesItemRoles(w, nil, "test-id", params)
			},
			skip: false,
		},
		{
			name: "TestPatchIdentitiesItemRolesFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentitiesService) {
				mockService.EXPECT().PatchIdentityRoles(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				patches, _ := json.Marshal(&resources.EntityEntitlementPatchRequestBody{})
				request := httptest.NewRequest(http.MethodPatch, "/identities/test-id/roles", bytes.NewReader(patches))
				h.PatchIdentitiesItemRoles(w, request, "test-id")
			},
			skip: true,
		},
	} {
		tt := test
		if tt.skip {
			continue
		}

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
			if err != nil {
				t.Errorf("Expected error to be nil, got %v", err)
			}

			response := new(resources.Response)

			if err := json.Unmarshal(data, response); err != nil {
				t.Errorf("Unexpected err while unmarshaling resonse, got: %v", err)
			}

			c.Assert(response.Status, qt.Equals, 500)
			c.Assert(response.Message, qt.Equals, "mock-error")
		})
	}
}

func TestPutIdentitiesItemFailureValidation(t *testing.T) {
	c := qt.New(t)

	expectedErrorResponse := resources.Response{
		Message: "Validation error: Identity ID from path does not match the Identity object",
		Status:  http.StatusBadRequest,
	}

	mockWriter := httptest.NewRecorder()
	sut := handler{}

	identity, _ := json.Marshal(&resources.Identity{Id: &mockIdentityId})
	request := httptest.NewRequest(http.MethodPut, "/identities/different-id", bytes.NewReader(identity))
	sut.PutIdentitiesItem(mockWriter, request, "different-id")

	result := mockWriter.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	response := new(resources.Response)

	if err := json.Unmarshal(data, &response); err != nil {
		t.Errorf("Unexpected err while unmarshaling resonse, got: %v", err)
	}

	c.Assert(result.StatusCode, qt.Equals, http.StatusBadRequest)
	c.Assert(response, qt.DeepEquals, &expectedErrorResponse)
}
