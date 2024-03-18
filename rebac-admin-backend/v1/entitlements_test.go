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

//go:generate mockgen -package interfaces -destination ./interfaces/mock_entitlements.go -source=./interfaces/entitlements.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Entitlements_Success(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntitlements := &resources.PaginatedResponse[resources.EntityEntitlement]{
		Data: []resources.EntityEntitlement{
			{
				EntitlementType: "mock-entl-type",
				EntityName:      "mock-entity-name",
				EntityType:      "mock-entity-type",
			},
		},
	}

	mockEntitlementsRaw := "mock-entitlements-raw-string"

	type EndpointTest struct {
		name             string
		setupServiceMock func(mockService *interfaces.MockEntitlementsService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
		expectedStatus   int
		expectedBody     any
	}

	tests := []EndpointTest{
		{
			name: "TestHandler_Entitlements_GetEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockEntitlementsService) {
				mockService.EXPECT().
					ListEntitlements(gomock.Any(), gomock.Eq(&resources.GetEntitlementsParams{})).
					Return(mockEntitlements, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/entitlements", nil)
				h.GetEntitlements(w, mockRequest, resources.GetEntitlementsParams{})
			},
			expectedStatus: http.StatusOK,
			expectedBody: resources.GetEntitlementsResponse{
				Data:   mockEntitlements.Data,
				Status: http.StatusOK,
			},
		},
		{
			name: "TestHandler_Entitlements_GetRawEntitlementsSuccess",
			setupServiceMock: func(mockService *interfaces.MockEntitlementsService) {
				mockService.EXPECT().
					RawEntitlements(gomock.Any()).
					Return(mockEntitlementsRaw, nil)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/entitlements/raw", nil)
				h.GetRawEntitlements(w, mockRequest)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mockEntitlementsRaw,
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockEntitlementsService := interfaces.NewMockEntitlementsService(ctrl)
			tt.setupServiceMock(mockEntitlementsService)

			sut := handler{Entitlements: mockEntitlementsService}

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

func TestHandler_Entitlements_ServiceBackendFailures(t *testing.T) {
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
		setupServiceMock func(mockService *interfaces.MockEntitlementsService)
		triggerFunc      func(h handler, w *httptest.ResponseRecorder)
	}

	tests := []EndpointTest{
		{
			name: "TestGetEntitlementsFailure",
			setupServiceMock: func(mockService *interfaces.MockEntitlementsService) {
				mockService.EXPECT().ListEntitlements(gomock.Any(), gomock.Any()).Return(nil, mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				mockRequest := httptest.NewRequest(http.MethodGet, "/entitlements", nil)
				h.GetEntitlements(w, mockRequest, resources.GetEntitlementsParams{})
			},
		},
		{
			name: "TestGetEntitlementsRawFailure",
			setupServiceMock: func(mockService *interfaces.MockEntitlementsService) {
				mockService.EXPECT().RawEntitlements(gomock.Any()).Return("", mockError)
			},
			triggerFunc: func(h handler, w *httptest.ResponseRecorder) {
				request := httptest.NewRequest(http.MethodGet, "/entitlements/raw", nil)
				h.GetRawEntitlements(w, request)
			},
		},
	}

	for _, test := range tests {
		tt := test
		c.Run(tt.name, func(c *qt.C) {
			mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)
			mockErrorResponseMapper.EXPECT().MapError(gomock.Any()).Return(&mockErrorResponse)

			mockEntitlementsService := interfaces.NewMockEntitlementsService(ctrl)
			tt.setupServiceMock(mockEntitlementsService)

			mockWriter := httptest.NewRecorder()
			sut := handler{
				Entitlements:            mockEntitlementsService,
				EntitlementsErrorMapper: mockErrorResponseMapper,
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
