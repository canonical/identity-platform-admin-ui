// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authorization

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

func TestIdentityConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/identities",
			input: input{method: http.MethodPost, endpoint: "/api/v0/identities"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/identities/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PUT /api/v0/identities/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/identities/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(IdentityConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestClientConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", CLIENT_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", CLIENT_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/clients",
			input: input{method: http.MethodPost, endpoint: "/api/v0/clients"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", CLIENT_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", CLIENT_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/clients/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PUT /api/v0/clients/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/clients/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/clients/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", CLIENT_TYPE, "id-1234")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(ClientConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v, expected %v", result, test.output)
			}
		})
	}
}

func TestProviderConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/idps",
			input: input{method: http.MethodPost, endpoint: "/api/v0/idps"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/idps/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v0/idps/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/idps/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/idps/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(ProviderConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestSchemeConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", SCHEME_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/schemas",
			input: input{method: http.MethodPost, endpoint: "/api/v0/schemas"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", SCHEME_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/schemas/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v0/schemas/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/schemas/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/schemas/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/schemas/default",
			input: input{method: http.MethodGet, endpoint: "/api/v0/schemas/default"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**")),
					},
				},
			},
		},
		{
			name:  "PUT /api/v0/schemas/default",
			input: input{method: http.MethodPut, endpoint: "/api/v0/schemas/default"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", SCHEME_TYPE, "**DEFAULT**")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(SchemeConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v, expected: %v", result, test.output)
			}
		})
	}
}

func TestRoleConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/roles/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v0/roles/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/roles/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/roles/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/roles/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/roles/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/roles/id-1234/identities/user-1",
			input: input{method: http.MethodPost, endpoint: "/api/v0/roles/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
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

			result := new(RoleConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestGroupConverterMapV0ReturnsPermissions(t *testing.T) {
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
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/groups",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/groups/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v0/groups/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "GET /api/v0/groups/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/identities", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "POST /api/v0/groups/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "POST /api/v0/groups/id-1234/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v0/groups/id-1234/roles", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/roles/viewer",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/roles", ID: "id-1234", RoleID: "viewer"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "viewer")},
			},
		},
		{
			name:  "DELETE /api/v0/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v0/groups/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "PATCH /api/v0/groups/id-1234/identities",
			input: input{method: http.MethodPatch, endpoint: "/api/v0/groups/id-1234/identities", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
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

			result := new(GroupConverter).MapV0(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestIdentityConverterMapV1ReturnsPermissions(t *testing.T) {
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
			name:  "GET /api/v1/identities",
			input: input{method: http.MethodGet, endpoint: "/api/v1/identities"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/identities",
			input: input{method: http.MethodPost, endpoint: "/api/v1/identities"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/identities/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v1/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PUT /api/v1/identities/id-1234",
			input: input{method: http.MethodPut, endpoint: "/api/v1/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/identities/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/identities/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", IDENTITY_TYPE, "id-1234")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(IdentityConverter).MapV1(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestProviderConverterMapV1ReturnsPermissions(t *testing.T) {
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
			name:   "GET /api/v1/authentication/providers",
			input:  input{method: http.MethodGet, endpoint: "/api/v1/authentication/providers"},
			output: []Permission{},
		},
		{
			name:  "GET /api/v1/authentication",
			input: input{method: http.MethodGet, endpoint: "/api/v1/authentication"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/authentication",
			input: input{method: http.MethodPost, endpoint: "/api/v1/authentication"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/authentication/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v1/authentication/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v1/authentication/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v1/authentication/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/authentication/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/authentication/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", PROVIDER_TYPE, "id-1234")),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.input.method, test.input.endpoint, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.input.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			result := new(ProviderConverter).MapV1(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestRoleConverterMapV1ReturnsPermissions(t *testing.T) {
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
			name:  "GET /api/v1/roles",
			input: input{method: http.MethodGet, endpoint: "/api/v1/roles"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v1/roles"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", ROLE_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/roles/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v1/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v1/roles/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v1/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/roles/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/roles/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/roles/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v1/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/roles/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/roles/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/roles/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v1/roles/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/roles/id-1234/identities/user-1",
			input: input{method: http.MethodPost, endpoint: "/api/v1/roles/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", "role:id-1234"),
					},
				},
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

			result := new(RoleConverter).MapV1(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}

func TestGroupConverterMapV1ReturnsPermissions(t *testing.T) {
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
			name:  "GET /api/v1/groups",
			input: input{method: http.MethodGet, endpoint: "/api/v1/groups"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/groups",
			input: input{method: http.MethodPost, endpoint: "/api/v1/groups"},
			output: []Permission{
				{
					Relation:   CAN_CREATE,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "__system__global"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("user:*", CAN_VIEW, fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, GLOBAL_ACCESS_OBJECT_NAME)),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/groups/id-1234",
			input: input{method: http.MethodGet, endpoint: "/api/v1/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "PATCH /api/v1/groups/id-1234",
			input: input{method: http.MethodPatch, endpoint: "/api/v1/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/groups/id-1234",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/groups/id-1234", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_DELETE,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "GET /api/v1/groups/id-1234/entitlements",
			input: input{method: http.MethodGet, endpoint: "/api/v1/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_VIEW,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/groups/id-1234/entitlements/can_view::role:1",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/groups/id-1234/entitlements", ID: "id-1234", EntitlementID: "can_view::role:1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/groups/id-1234/identities", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "POST /api/v1/groups/id-1234/entitlements",
			input: input{method: http.MethodPost, endpoint: "/api/v1/groups/id-1234/entitlements", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "POST /api/v1/groups/id-1234/roles",
			input: input{method: http.MethodPost, endpoint: "/api/v1/groups/id-1234/roles", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
			},
		},
		{
			name:  "DELETE /api/v1/groups/id-1234/roles/viewer",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/groups/id-1234/roles", ID: "id-1234", RoleID: "viewer"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, "viewer")},
			},
		},
		{
			name:  "DELETE /api/v1/groups/id-1234/identities/user-1",
			input: input{method: http.MethodDelete, endpoint: "/api/v1/groups/id-1234/identities/user-1", ID: "id-1234", IdentityID: "user-1"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
				{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, "user-1")},
			},
		},
		{
			name:  "PATCH /api/v1/groups/id-1234/identities",
			input: input{method: http.MethodPatch, endpoint: "/api/v1/groups/id-1234/identities", ID: "id-1234"},
			output: []Permission{
				{
					Relation:   CAN_EDIT,
					ResourceID: fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234"),
					ContextualTuples: []openfga.Tuple{
						*openfga.NewTuple("privileged:superuser", "privileged", fmt.Sprintf("%s:%s", GROUP_TYPE, "id-1234")),
					},
				},
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

			result := new(GroupConverter).MapV1(r)

			if !reflect.DeepEqual(result, test.output) {
				t.Errorf("Map returned %v", result)
			}
		})
	}
}
