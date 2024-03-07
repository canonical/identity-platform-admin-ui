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
	mockIdentityProviderName = "MockProviderName"
	mockIdentityProviderId   = "test-id"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_identity_providers.go -source=./interfaces/identity_providers.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_IdP_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIDPObject := resources.IdentityProvider{
		Id:   &mockIdentityProviderId,
		Name: &mockIdentityProviderName,
	}

	mockAvailableIDPObject := resources.AvailableIdentityProvider{
		Id:   mockIdentityProviderId,
		Name: &mockIdentityProviderName,
	}

	mockIDPs := []resources.IdentityProvider{mockIDPObject}
	mockAvailableIDPs := []resources.AvailableIdentityProvider{mockAvailableIDPObject}

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockIdentityProvidersService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_IdP_GetAvailableIdentityProvidersSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				params := resources.GetAvailableIdentityProvidersParams{}
				mockService.EXPECT().
					ListAvailableIdentityProviders(gomock.Any(), gomock.Eq(&params)).
					Return(&resources.AvailableIdentityProviders{
						Data: []resources.AvailableIdentityProvider{mockAvailableIDPObject},
					}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/authentication/providers", nil)
				h.GetAvailableIdentityProviders(w, mockRequest, resources.GetAvailableIdentityProvidersParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetAvailableIdentityProvidersResponse{
				Data:   mockAvailableIDPs,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_IdP_GetIdentityProvidersSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().
					ListIdentityProviders(gomock.Any(), gomock.Any()).
					Return(&resources.IdentityProviders{Data: mockIDPs}, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				params := resources.GetIdentityProvidersParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/authentication", nil)
				h.GetIdentityProviders(w, mockRequest, params)
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetIdentityProvidersResponse{
				Data:   mockIDPs,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_IdP_PostIdentityProvidersSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().
					RegisterConfiguration(gomock.Any(), gomock.Eq(&mockIDPObject)).
					Return(&mockIDPObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				idpBody, _ := json.Marshal(mockIDPObject)
				mockRequest := httptest.NewRequest(http.MethodPost, "/authentication", bytes.NewReader(idpBody))
				h.PostIdentityProviders(w, mockRequest)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   mockIDPObject,
		},
		{
			name: "TestHandler_IdP_DeleteIdentityProvidersItemSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().
					DeleteConfiguration(gomock.Any(), gomock.Eq(mockIdentityProviderId)).
					Return(true, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), nil)
				h.DeleteIdentityProvidersItem(w, mockRequest, mockIdentityProviderId)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "TestHandler_IdP_GetIdentityProvidersItemSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().
					GetConfiguration(gomock.Any(), gomock.Eq(mockIdentityProviderId)).
					Return(&mockIDPObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), nil)
				h.GetIdentityProvidersItem(w, mockRequest, mockIdentityProviderId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockIDPObject,
		},
		{
			name: "TestHandler_IdP_PutIdentityProvidersItemSuccess",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().
					UpdateConfiguration(gomock.Any(), gomock.Eq(&mockIDPObject)).
					Return(&mockIDPObject, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				idpBody, _ := json.Marshal(mockIDPObject)
				mockRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), bytes.NewReader(idpBody))
				h.PutIdentityProvidersItem(w, mockRequest, mockIdentityProviderId)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockIDPObject,
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockIDPService := interfaces.NewMockIdentityProvidersService(ctrl)
			tt.setupServiceMock(mockIDPService)

			sut := handler{IdentityProviders: mockIDPService}

			mockWriter := httptest.NewRecorder()
			tt.triggerFunc(sut, mockWriter)

			result := mockWriter.Result()
			defer result.Body.Close()

			c.Assert(result.StatusCode, qt.Equals, tt.expectedStatus)

			body, err := io.ReadAll(result.Body)
			c.Assert(err, qt.IsNil)

			c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))

			if tt.expectedBody != nil {
				c.Assert(string(body), qt.JSONEquals, tt.expectedBody)
			} else {
				c.Assert(len(body), qt.Equals, 0)
			}
		})
	}

}

func TestHandler_IdP_ValidationErrors(t *testing.T) {
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
			name: "TestPostIdentityProvidersFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/authentication", bytes.NewReader(invalidRequestBody))
				h.PostIdentityProviders(w, req)
			},
		},
		{
			name: "TestPutIdentityProvidersItemFailureInvalidRequest",
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), bytes.NewReader(invalidRequestBody))
				h.PutIdentityProvidersItem(w, req, mockIdentityProviderId)
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

func TestHandler_IdP_ServiceBackendFailures(t *testing.T) {
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
		setupServiceMock func(mockService *interfaces.MockIdentityProvidersService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		skip             bool
	}

	tests := []EndpointTest{
		{
			name: "GetAvailableIdentityProvidersFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().ListAvailableIdentityProviders(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockParams := resources.GetAvailableIdentityProvidersParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/authentication/providers", nil)
				h.GetAvailableIdentityProviders(w, mockRequest, mockParams)
			},
		},
		{
			name: "TestGetIdentityProvidersFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().ListIdentityProviders(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := httptest.NewRequest(http.MethodGet, "/authentication", nil)
				h.GetIdentityProviders(w, request, resources.GetIdentityProvidersParams{})
			},
		},
		{
			name: "TestPostIdentityProvidersFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().RegisterConfiguration(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				idp, _ := json.Marshal(&resources.IdentityProvider{})
				mockRequest := httptest.NewRequest(http.MethodPost, "/authentication", bytes.NewReader(idp))
				h.PostIdentityProviders(w, mockRequest)
			},
		},
		{
			name: "TestDeleteIdentityProvidersItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().DeleteConfiguration(gomock.Any(), gomock.Any()).Return(false, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), nil)
				h.DeleteIdentityProvidersItem(w, mockRequest, mockIdentityProviderId)
			},
		},
		{
			name: "TestGetIdentityProvidersItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().GetConfiguration(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), nil)
				h.GetIdentityProvidersItem(w, request, mockIdentityProviderId)
			},
		},
		{
			name: "TestPutIdentityProvidersItemFailure",
			setupServiceMock: func(mockService *interfaces.MockIdentityProvidersService) {
				mockService.EXPECT().UpdateConfiguration(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				idp, _ := json.Marshal(&resources.IdentityProvider{Id: &mockIdentityProviderId})
				mockRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/authentication/%s", mockIdentityProviderId), bytes.NewReader(idp))
				h.PutIdentityProvidersItem(w, mockRequest, mockIdentityProviderId)
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockIDPService := interfaces.NewMockIdentityProvidersService(ctrl)
			tt.setupServiceMock(mockIDPService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				IdentityProviders:            mockIDPService,
				IdentityProvidersErrorMapper: mockErrorResponseMapper,
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
