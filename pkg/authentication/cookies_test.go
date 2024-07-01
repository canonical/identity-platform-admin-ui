// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package authentication -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go

func findCookie(name string, cookies []*http.Cookie) (*http.Cookie, bool) {
	for _, cookie := range cookies {
		if name == cookie.Name {
			return cookie, true
		}
	}

	return nil, false
}

func TestAuthCookieManager_ClearNonceCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "nonce"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.ClearNonceCookie(mockResponse)

	c, _ := findCookie("nonce", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear nonce cookie")
	}
}

func TestAuthCookieManager_ClearStateCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "state"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.ClearStateCookie(mockResponse)

	c, _ := findCookie("state", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear state cookie")
	}
}

func TestAuthCookieManager_ClearIDTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "id-token"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.ClearIDTokenCookie(mockResponse)

	c, _ := findCookie("id-token", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear id token cookie")
	}
}

func TestAuthCookieManager_AccessIDTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "access-token"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.ClearAccessTokenCookie(mockResponse)

	c, _ := findCookie("access-token", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear access token cookie")
	}
}

func TestAuthCookieManager_ClearRefreshTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "refresh-token"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.ClearRefreshTokenCookie(mockResponse)

	c, _ := findCookie("refresh-token", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear refresh token cookie")
	}
}

func TestAuthCookieManager_GetNonceCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-nonce").Return("mock-nonce", nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "nonce", Value: "mock-nonce"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetNonceCookie(mockRequest)

	if cookie != "mock-nonce" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetNonceCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't get cookie %s, %v", "nonce", gomock.Any()).Times(1)
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)

	manager := NewAuthCookieManager(5, 5, nil, mockLogger)
	cookie := manager.GetNonceCookie(mockRequest)

	if cookie != "" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetNonceCookieDecryptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't decrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-nonce").Return("", mockError)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "nonce", Value: "mock-nonce"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetNonceCookie(mockRequest)

	if cookie != "" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetStateCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-state").Return("mock-state", nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "state", Value: "mock-state"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetStateCookie(mockRequest)

	if cookie != "mock-state" {
		t.Fatal("state cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetStateCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't get cookie %s, %v", "state", gomock.Any()).Times(1)
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)

	manager := NewAuthCookieManager(5, 5, nil, mockLogger)
	cookie := manager.GetStateCookie(mockRequest)

	if cookie != "" {
		t.Fatal("state cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetStateCookieDecryptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't decrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-state").Return("", mockError)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "state", Value: "mock-state"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetStateCookie(mockRequest)

	if cookie != "" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetIDTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-idtoken").Return("mock-idtoken", nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "id-token", Value: "mock-idtoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetIDTokenCookie(mockRequest)

	if cookie != "mock-idtoken" {
		t.Fatal("id token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetIDTokenCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't get cookie %s, %v", "id-token", gomock.Any()).Times(1)
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)

	manager := NewAuthCookieManager(5, 5, nil, mockLogger)
	cookie := manager.GetIDTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("id-token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetIDTokenCookieDecryptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't decrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-idtoken").Return("", mockError)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "id-token", Value: "mock-idtoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetIDTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("id token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetAccessTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-accesstoken").Return("mock-accesstoken", nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "access-token", Value: "mock-accesstoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetAccessTokenCookie(mockRequest)

	if cookie != "mock-accesstoken" {
		t.Fatal("access token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetAccessTokenCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't get cookie %s, %v", "access-token", gomock.Any()).Times(1)
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)

	manager := NewAuthCookieManager(5, 5, nil, mockLogger)
	cookie := manager.GetAccessTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("access-token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetAccessTokenCookieDecryptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't decrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-accesstoken").Return("", mockError)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "access-token", Value: "mock-accesstoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetAccessTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("access token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetRefreshTokenCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-refreshtoken").Return("mock-refreshtoken", nil)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "refresh-token", Value: "mock-refreshtoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetRefreshTokenCookie(mockRequest)

	if cookie != "mock-refreshtoken" {
		t.Fatal("refresh token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetRefreshTokenCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't get cookie %s, %v", "refresh-token", gomock.Any()).Times(1)
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)

	manager := NewAuthCookieManager(5, 5, nil, mockLogger)
	cookie := manager.GetRefreshTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("refresh token cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetRefreshTokenCookieDecryptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't decrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Decrypt("mock-refreshtoken").Return("", mockError)

	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "refresh-token", Value: "mock-refreshtoken"})

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	cookie := manager.GetRefreshTokenCookie(mockRequest)

	if cookie != "" {
		t.Fatal("refresh token cookie value does not match expected")
	}
}

func TestAuthCookieManager_SetNonceCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Encrypt("mock-nonce").Return("mock-nonce", nil)

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.SetNonceCookie(mockResponse, "mock-nonce")

	c, found := findCookie("nonce", mockResponse.Result().Cookies())

	if !found {
		t.Fatal("did not set nonce cookie")
	}

	if c.Value != "mock-nonce" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_SetNonceCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't encrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Encrypt("mock-nonce").Return("", mockError)

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.SetNonceCookie(mockResponse, "mock-nonce")
}

func TestAuthCookieManager_SetStateCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockLogger := NewMockLoggerInterface(ctrl)
	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Encrypt("mock-state").Return("mock-state", nil)

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.SetStateCookie(mockResponse, "mock-state")

	c, found := findCookie("state", mockResponse.Result().Cookies())

	if !found {
		t.Fatal("did not set state cookie")
	}

	if c.Value != "mock-state" {
		t.Fatal("state cookie value does not match expected")
	}
}

func TestAuthCookieManager_SetStateCookieFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockError := errors.New("mock-error")

	mockLogger := NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf("can't encrypt cookie value, %v", mockError).Times(1)

	mockEncrypt := NewMockEncryptInterface(ctrl)
	mockEncrypt.EXPECT().Encrypt("mock-state").Return("", mockError)

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager(5, 5, mockEncrypt, mockLogger)
	manager.SetStateCookie(mockResponse, "mock-state")
}
