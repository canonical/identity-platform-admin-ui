// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package authorization

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestIdentityConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method   string
		endpoint string
		ID       string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/identities",
			input: input{method: http.MethodGet, endpoint: "/api/v0/identities"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/identities",
			input: input{method: http.MethodPost, endpoint: "/api/v0/identities"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/identities/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")},
			},
		},
		{
			name:  "PUT /api/v0/identities/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/identities/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(IdentityConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestClientConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method   string
		endpoint string
		ID       string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/clients",
			input: input{method: http.MethodGet, endpoint: "/api/v0/clients"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/clients",
			input: input{method: http.MethodPost, endpoint: "/api/v0/clients"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/clients/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")},
			},
		},
		{
			name:  "PUT /api/v0/clients/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/clients/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(ClientConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestProviderConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method   string
		endpoint string
		ID       string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/idps",
			input: input{method: http.MethodGet, endpoint: "/api/v0/idps"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/idps",
			input: input{method: http.MethodPost, endpoint: "/api/v0/idps"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/idps/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")},
			},
		},
		{
			name:  "PATCH /api/v0/idps/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/idps/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(ProviderConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestRuleConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method   string
		endpoint string
		ID       string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/rules",
			input: input{method: http.MethodGet, endpoint: "/api/v0/rules"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/rules",
			input: input{method: http.MethodPost, endpoint: "/api/v0/rules"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/rules/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/rules/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "id-1234")},
			},
		},
		{
			name:  "PUT /api/v0/rules/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v0/rules/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/rules/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/rules/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", RULE_TYPE, "id-1234")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(RuleConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestSchemeConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method   string
		endpoint string
		ID       string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/schemas",
			input: input{method: http.MethodGet, endpoint: "/api/v0/schemas"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/schemas",
			input: input{method: http.MethodPost, endpoint: "/api/v0/schemas"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/schemas/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")},
			},
		},
		{
			name:  "PATCH /api/v0/schemas/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/schemas/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")},
			},
		},
		{
			name:  "GET /api/v0/schemas/default",
			input: input{method: http.MethodGet, endpoint: "/api/v0/schemas/default"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**")},
			},
		},
		{
			name:  "PUT /api/v0/schemas/default",
			input: input{method: http.MethodPut, endpoint: "/api/v0/schemas/default"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(SchemeConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestRoleConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method        string
		endpoint      string
		ID            string
		EntitlementID string
		IdentityID    string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/roles",
			input: input{method: http.MethodGet, endpoint: "/api/v0/roles"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/roles/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "PATCH /api/v0/roles/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/roles/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "GET /api/v0/roles/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/roles/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "POST /api/v0/roles/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
			},
		},
		{
			name:  "POST /api/v0/roles/id-1234/identities/user-1",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234")},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)
			rctx.URLParams.Add("e_id", test.input.EntitlementID)
			rctx.URLParams.Add("i_id", test.input.IdentityID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(RoleConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestGroupConverterMapReturnsPermissions(t *testing.T) {
	type input struct {
		method        string
		endpoint      string
		ID            string
		RoleID        string
		EntitlementID string
		IdentityID    string
	}

	tests := []struct {
		name   string
		input  input
		output []Permission
	}{
		{
			name:  "GET /api/v0/groups",
			input: input{method: http.MethodGet, endpoint: "/api/v0/groups"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "global")},
			},
		},
		{
			name:  "POST /api/v0/groups",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups"},
			output: []Permission{
				{Relation: CAN_CREATE, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "global")},
			},
		},
		{
			name:  "GET /api/v0/groups/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "PATCH /api/v0/groups/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_DELETE, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "GET /api/v0/groups/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/identities", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "POST /api/v0/groups/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "POST /api/v0/groups/id-1234/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups/id-1234/roles", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/roles/viewer",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/roles", ID: "id-1234", RoleID: "viewer"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "viewer")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "PATCH /api/v0/groups/id-1234/identities",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/groups/id-1234/identities", ID: "id-1234"},
			output: []Permission{
				{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)
			rctx.URLParams.Add("r_id", test.input.RoleID)
			rctx.URLParams.Add("e_id", test.input.EntitlementID)
			rctx.URLParams.Add("i_id", test.input.IdentityID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(GroupConverter).Map(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}
