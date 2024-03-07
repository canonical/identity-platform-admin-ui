// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
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

//go:generate mockgen -package interfaces -destination ./interfaces/mock_capabilities.go -source=./interfaces/capabilities.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

func TestHandler_Capabilities_GetCapabilitiesSuccess(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCapabilitiesService := interfaces.NewMockCapabilitiesService(ctrl)

	mockCapabilities := []resources.Capability{
		{
			Endpoint: "/capabilities",
			Methods:  []resources.CapabilityMethods{"GET"},
		},
		{
			Endpoint: "/identities",
			Methods:  []resources.CapabilityMethods{"GET", "POST", "DELETE", "PUT"},
		},
		{
			Endpoint: "/roles",
			Methods:  []resources.CapabilityMethods{"GET", "POST", "DELETE", "PUT"},
		},
	}

	mockReturnCapabilities := resources.Capabilities{Data: mockCapabilities}
	mockCapabilitiesService.EXPECT().ListCapabilities(gomock.Any()).Return(&mockReturnCapabilities, nil)

	expectedResponse := resources.GetCapabilitiesResponse{
		Data:   mockCapabilities,
		Status: http.StatusOK,
	}

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, "/capabilities", nil)

	sut := handler{Capabilities: mockCapabilitiesService}
	sut.GetCapabilities(mockWriter, mockRequest)

	result := mockWriter.Result()
	defer result.Body.Close()

	responseBody, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(string(responseBody), qt.JSONEquals, &expectedResponse)
}

func TestHandler_Capabilities_GetCapabilitiesFailure(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCapabilitiesService := interfaces.NewMockCapabilitiesService(ctrl)
	mockErrorResponseMapper := NewMockErrorResponseMapper(ctrl)

	mockErrorResponse := resources.Response{
		Message: "mock-error",
		Status:  http.StatusInternalServerError,
	}

	mockError := errors.New("test-error")

	mockCapabilitiesService.EXPECT().ListCapabilities(gomock.Any()).Return(nil, mockError)
	mockErrorResponseMapper.EXPECT().MapError(gomock.Eq(mockError)).Return(&mockErrorResponse)

	sut := handler{
		Capabilities:            mockCapabilitiesService,
		CapabilitiesErrorMapper: mockErrorResponseMapper,
	}

	mockWriter := httptest.NewRecorder()
	mockRequest := httptest.NewRequest(http.MethodGet, "/capabilities", nil)

	sut.GetCapabilities(mockWriter, mockRequest)

	result := mockWriter.Result()
	defer result.Body.Close()

	c.Assert(result.StatusCode, qt.Equals, http.StatusInternalServerError)

	data, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)

	c.Assert(err, qt.IsNil, qt.Commentf("Unexpected err while unmarshaling resonse, got: %v", err))
	c.Assert(result.StatusCode, qt.Equals, http.StatusInternalServerError)
	c.Assert(data, qt.JSONEquals, mockErrorResponse)

}
