// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package authorization

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package authorization -destination ./mock_monitor.go -source=../monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authorization -destination ./mock_logger.go -source=../logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package authorization -destination ./mock_interfaces.go -source=./interfaces.go

type API struct{}

func (a *API) RegisterEndpoints(router *chi.Mux) {
	router.Get("/api/v0/identities", a.handleAll)
	router.Get("/api/v0/identities/1", a.handleAll)
	router.Post("/api/v0/clients", a.handleAll)
	router.Get("/api/v0/idps/github", a.handleAll)
	router.Delete("/api/v0/rules/1", a.handleAll)
	router.Patch("/api/v0/schemas/x", a.handleAll)
	router.Post("/api/v0/roles/viewer/identities/1", a.handleAll)
	router.Get("/api/v0/allow", a.handleAll)
	router.Get("/api/v0/forbidden", a.handleAll)
}

func (a *API) handleAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func TestMiddlewareAuthorize(t *testing.T) {
	type input struct {
		method     string
		endpoint   string
		ID         string
		IdentityID string
	}

	tests := []struct {
		name   string
		input  input
		expect []Permission
		output bool
	}{
		{
			name:   "GET /api/v0/allow",
			input:  input{method: http.MethodGet, endpoint: "/api/v0/allow"},
			expect: []Permission{},
			output: true,
		},
		{
			name:  "GET /api/v0/identities/1",
			input: input{method: http.MethodGet, endpoint: "/api/v0/identities/1", ID: "1"},
			expect: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "1")},
			},
			output: false,
		},
		{
			name:  "GET /api/v0/identities",
			input: input{method: http.MethodGet, endpoint: "/api/v0/identities"},
			expect: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "global")},
			},
			output: true,
		},
		{
			name:  "GET /api/v0/idps/github",
			input: input{method: http.MethodGet, endpoint: "/api/v0/idps/github", ID: "github"},
			expect: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "github")},
			},
			output: true,
		},
		{
			name:  "PATCH /api/v0/schemas/x",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/schemas/x", ID: "x"},
			expect: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "x")},
			},
			output: true,
		},
		{
			name:  "DELETE /api/v0/rules/1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/rules/1", ID: "1"},
			expect: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "1")},
			},
			output: true,
		},
		{
			name:  "POST /api/v0/roles/viewer/identities/1",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles/viewer/identities/1", ID: "viewer", IdentityID: "1"},
			expect: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "viewer")},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "1")},
			},
			output: true,
		},
		{
			name:  "POST /api/v0/clients",
			input: input{method: http.MethodPost, endpoint: "/api/v0/clients"},
			expect: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "global")},
			},
			output: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMonitor := NewMockMonitorInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockAuthorizer := NewMockAuthorizerInterface(ctrl)

			router := chi.NewMux().With(
				NewMiddleware(mockAuthorizer, mockMonitor, mockLogger).Authorize(),
			).(*chi.Mux)

			new(API).RegisterEndpoints(router)

			calls := []*gomock.Call{}

			mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			calls = append(calls, mockAuthorizer.EXPECT().Check(gomock.Any(), "user:admin", "admin", ADMIN_PRIVILEGE).Times(1).Return(false, nil))
			for _, check := range test.expect {
				calls = append(
					calls,
					mockAuthorizer.EXPECT().Check(gomock.Any(), gomock.Any(), check.Relation, check.ResourceID).Times(1).Return(test.output, nil),
				)
			}

			gomock.InAnyOrder(calls)

			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)
			r.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)
			rctx.URLParams.Add("i_id", test.input.IdentityID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if !test.output && w.Result().StatusCode != http.StatusForbidden {
				t.Fatalf("expected HTTP status code 403 got %v", w.Result().StatusCode)
			}

			if test.output && w.Result().StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP status code 200 got %v", w.Result().StatusCode)
			}

		})
	}
}

func TestMiddlewareAuthorizeUseTokenHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMonitor := NewMockMonitorInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)
	mockAuthorizer := NewMockAuthorizerInterface(ctrl)

	router := chi.NewMux().With(
		NewMiddleware(mockAuthorizer, mockMonitor, mockLogger).Authorize(),
	).(*chi.Mux)

	new(API).RegisterEndpoints(router)

	testUser := "test-user"

	calls := []*gomock.Call{}

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

	calls = append(
		calls,
		mockAuthorizer.EXPECT().Check(gomock.Any(), fmt.Sprintf("user:%s", testUser), "admin", ADMIN_PRIVILEGE).Times(1).Return(false, nil),
		mockAuthorizer.EXPECT().Check(gomock.Any(), gomock.Any(), CAN_VIEW, fmt.Sprintf("%s:%s", IDENTITY_TYPE, "global")).Times(1).Return(true, nil),
	)

	gomock.InAnyOrder(calls)

	r := httptest.NewRequest(http.MethodGet, "/api/v0/identities", nil)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(TOKEN_HEADER, base64.RawStdEncoding.EncodeToString([]byte(testUser)))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code 200 got %v", w.Result().StatusCode)
	}
}
