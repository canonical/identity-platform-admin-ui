// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "nonce"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager()
	manager.ClearNonceCookie(mockResponse)

	c, _ := findCookie("nonce", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear nonce cookie")
	}
}

func TestAuthCookieManager_ClearStateCookie(t *testing.T) {
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "state"})

	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager()
	manager.ClearStateCookie(mockResponse)

	c, _ := findCookie("state", mockResponse.Result().Cookies())

	if c.Expires != epoch {
		t.Fatal("did not clear state cookie")
	}
}

func TestAuthCookieManager_GetNonceCookie(t *testing.T) {
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "nonce", Value: "mock-nonce"})

	manager := NewAuthCookieManager()
	cookie := manager.GetNonceCookie(mockRequest)

	if cookie != "mock-nonce" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_GetStateCookie(t *testing.T) {
	mockRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mockRequest.AddCookie(&http.Cookie{Name: "state", Value: "mock-state"})

	manager := NewAuthCookieManager()
	cookie := manager.GetStateCookie(mockRequest)

	if cookie != "mock-state" {
		t.Fatal("state cookie value does not match expected")
	}
}

func TestAuthCookieManager_SetNonceCookie(t *testing.T) {
	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager()
	manager.SetNonceCookie(mockResponse, "mock-nonce", time.Minute)

	c, found := findCookie("nonce", mockResponse.Result().Cookies())

	if !found {
		t.Fatal("did not set nonce cookie")
	}

	if c.Value != "mock-nonce" {
		t.Fatal("nonce cookie value does not match expected")
	}
}

func TestAuthCookieManager_SetStateCookie(t *testing.T) {
	mockResponse := httptest.NewRecorder()

	manager := NewAuthCookieManager()
	manager.SetStateCookie(mockResponse, "mock-state", time.Minute)

	c, found := findCookie("state", mockResponse.Result().Cookies())

	if !found {
		t.Fatal("did not set state cookie")
	}

	if c.Value != "mock-state" {
		t.Fatal("state cookie value does not match expected")
	}
}
