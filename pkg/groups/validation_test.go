// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"

	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

func TestNeedsValidation(t *testing.T) {
	p := new(PayloadValidator)
	p.validator = validation.NewValidator()

	for _, tt := range []struct {
		name           string
		req            *http.Request
		expectedResult bool
	}{
		{
			name:           http.MethodPost,
			req:            httptest.NewRequest(http.MethodPost, "/", nil),
			expectedResult: true,
		},
		{
			name:           http.MethodPut,
			req:            httptest.NewRequest(http.MethodPut, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodPatch,
			req:            httptest.NewRequest(http.MethodPatch, "/", nil),
			expectedResult: true,
		},
		{
			name:           http.MethodGet,
			req:            httptest.NewRequest(http.MethodGet, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodDelete,
			req:            httptest.NewRequest(http.MethodDelete, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodConnect,
			req:            httptest.NewRequest(http.MethodConnect, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodHead,
			req:            httptest.NewRequest(http.MethodHead, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodTrace,
			req:            httptest.NewRequest(http.MethodTrace, "/", nil),
			expectedResult: false,
		},
		{
			name:           http.MethodOptions,
			req:            httptest.NewRequest(http.MethodOptions, "/", nil),
			expectedResult: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := p.NeedsValidation(tt.req)

			if result != tt.expectedResult {
				t.Fatalf("Result doesn't match expected one, obtained %t instead of %t", result, tt.expectedResult)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	p := new(PayloadValidator)
	p.apiKey = "groups"
	p.validator = validation.NewValidator()
	_ = p.validator.RegisterValidation("notblank", validators.NotBlank)

	for _, tt := range []struct {
		name           string
		method         string
		endpoint       string
		body           func() []byte
		expectedResult validator.ValidationErrors
		expectedError  error
	}{
		{
			name:     "CreateGroup",
			method:   http.MethodPost,
			endpoint: "",
			body: func() []byte {
				r := new(Group)
				r.Name = "mock-name"

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:     "UpdateGroup",
			method:   http.MethodPatch,
			endpoint: "/mock-id",
			body: func() []byte {
				id := "mock-id"
				r := new(Group)
				r.ID = id

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:     "AssignRoles",
			method:   http.MethodPost,
			endpoint: "/mock-id/roles",
			body: func() []byte {
				r := new(UpdateRolesRequest)
				r.Roles = []string{
					"viewer", "writer",
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:     "AssignPermissions",
			method:   http.MethodPatch,
			endpoint: "/mock-id/entitlements",
			body: func() []byte {
				r := new(UpdatePermissionsRequest)
				r.Permissions = []Permission{
					{
						Relation: "mock-relation",
						Object:   "mock-object",
					},
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:     "AssignIdentities",
			method:   http.MethodPatch,
			endpoint: "/mock-id/identities",
			body: func() []byte {
				r := new(UpdateIdentitiesRequest)
				r.Identities = []string{
					"mock-identity",
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:     "NoMatch",
			method:   http.MethodPost,
			endpoint: "no-match-endpoint",
			body: func() []byte {
				return nil
			},
			expectedResult: nil,
			expectedError:  validation.NoMatchError(p.apiKey),
		},
		{
			name:     "CreateGroupFailure",
			method:   http.MethodPost,
			endpoint: "",
			body: func() []byte {
				r := new(Group)

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: validator.ValidationErrors{},
			expectedError:  nil,
		},
		{
			name:     "UpdateGroupFailure",
			method:   http.MethodPatch,
			endpoint: "/mock-id",
			body: func() []byte {
				r := new(Group)

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: validator.ValidationErrors{},
			expectedError:  nil,
		},
		{
			name:     "AssignRolesFailure",
			method:   http.MethodPost,
			endpoint: "/mock-id/roles",
			body: func() []byte {
				r := new(UpdateRolesRequest)
				r.Roles = []string{
					"viewer", "",
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: validator.ValidationErrors{},
			expectedError:  nil,
		},
		{
			name:     "AssignPermissionsFailure",
			method:   http.MethodPatch,
			endpoint: "/mock-id/entitlements",
			body: func() []byte {
				r := new(UpdatePermissionsRequest)
				r.Permissions = []Permission{
					{
						Relation: "",
						Object:   "mock-object",
					},
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: validator.ValidationErrors{},
			expectedError:  nil,
		},
		{
			name:     "AssignIdentitiesFailure",
			method:   http.MethodPatch,
			endpoint: "/mock-id/identities",
			body: func() []byte {
				r := new(UpdateIdentitiesRequest)
				r.Identities = []string{
					"",
				}

				marshal, _ := json.Marshal(r)
				return marshal
			},
			expectedResult: validator.ValidationErrors{},
			expectedError:  nil,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := p.Validate(context.TODO(), tt.method, tt.endpoint, tt.body())

			if err != nil {
				if tt.expectedError == nil {
					t.Fatalf("Unexpected error for '%s'", tt.name)
				}

				if err.Error() != tt.expectedError.Error() {
					t.Fatalf("Returned error doesn't match expected, obtained '%v' instead of '%v'", err, tt.expectedError)
				}

				return
			}

			if result != nil {
				if tt.expectedResult == nil {
					t.Fatalf("Unexpected result for '%s'", tt.name)
				}

				if errors.Is(result, tt.expectedResult) {
					t.Fatalf("Returned validation errors don't match expected, obtained '%v' instead of '%v'", result, tt.expectedResult)
				}

				return
			}
		})
	}
}
