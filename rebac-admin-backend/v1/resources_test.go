// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_resources.go -source=./interfaces/resources.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Resources_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResources := resources.PaginatedResponse[resources.Resource]{
		Data: []resources.Resource{{
			Id:   "resource-id",
			Name: "resource-name",
			Entity: resources.Entity{
				Id:   "entity-id",
				Type: "entity-type",
			},
		}},
	}

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockResourcesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Resources_ListResourcesSuccess",
			setupServiceMock: func(mockService *interfaces.MockResourcesService) {
				mockService.EXPECT().
					ListResources(gomock.Any(), gomock.Eq(&resources.GetResourcesParams{})).
					Return(&mockResources, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/resources", nil)
				h.GetResources(w, mockRequest, resources.GetResourcesParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetResourcesResponse{
				Data:   mockResources.Data,
				Status: http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockResourcesService := interfaces.NewMockResourcesService(ctrl)
			tt.setupServiceMock(mockResourcesService)

			sut := handler{Resources: mockResourcesService}

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

func TestHandler_Resources_ServiceBackendFailures(t *testing.T) {
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
		setupServiceMock func(mockService *interfaces.MockResourcesService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestGetResourcesFailure",
			setupServiceMock: func(mockService *interfaces.MockResourcesService) {
				mockService.EXPECT().ListResources(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockParams := resources.GetResourcesParams{}
				mockRequest := httptest.NewRequest(http.MethodGet, "/resources", nil)
				h.GetResources(w, mockRequest, mockParams)
			},
		},
	}
	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockResourcesService := interfaces.NewMockResourcesService(ctrl)
			tt.setupServiceMock(mockResourcesService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				Resources:            mockResourcesService,
				ResourcesErrorMapper: mockErrorResponseMapper,
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
