// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package authorization

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// these constants relate directly to the authorization model types
const (
	IDENTITY_TYPE = "identity"
	CLIENT_TYPE   = "client"
	PROVIDER_TYPE = "provider"
	RULE_TYPE     = "rule"
	SCHEME_TYPE   = "scheme"
	ROLE_TYPE     = "role"
	GROUP_TYPE    = "group"

	CAN_VIEW   = "can_view"
	CAN_EDIT   = "can_edit"
	CAN_CREATE = "can_create"
	CAN_DELETE = "can_delete"
)

type Permission struct {
	Relation   string
	ResourceID string
}

func relation(r *http.Request) string {
	switch r.Method {
	case http.MethodGet:
		return CAN_VIEW
	case http.MethodPost:
		return CAN_CREATE
	case http.MethodPut, http.MethodPatch:
		return CAN_EDIT
	case http.MethodDelete:
		return CAN_DELETE
	default:
		return CAN_VIEW
	}
}

type IdentityConverter struct{}

func (c IdentityConverter) TypeName() string {
	return IDENTITY_TYPE
}

func (c IdentityConverter) Map(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")

	if id == "" {
		// global is used in place of `*`, reason is OpenFGA has a special meaning for `*`
		// see [public access](https://openfga.dev/docs/modeling/public-access)
		id = "global"
	}

	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), id)},
	}
}

type ClientConverter struct{}

func (c ClientConverter) TypeName() string {
	return CLIENT_TYPE
}

func (c ClientConverter) Map(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")

	if id == "" {
		id = "global"
	}
	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), id)},
	}
}

type ProviderConverter struct{}

func (c ProviderConverter) TypeName() string {
	return PROVIDER_TYPE
}

func (c ProviderConverter) Map(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")

	if id == "" {
		id = "global"
	}
	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), id)},
	}
}

type RuleConverter struct{}

func (c RuleConverter) TypeName() string {
	return RULE_TYPE
}

func (c RuleConverter) Map(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")

	if id == "" {
		id = "global"
	}
	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), id)},
	}
}

type SchemeConverter struct{}

func (c SchemeConverter) TypeName() string {
	return SCHEME_TYPE
}

func (c SchemeConverter) Map(r *http.Request) []Permission {
	// TODO @shipperizer let's make sure this is a good way to codify the
	// default schema API
	if r.URL.Path == "/api/v0/schemas/default" {
		return []Permission{
			{
				Relation:   relation(r),
				ResourceID: fmt.Sprintf("%s:**DEFAULT**", c.TypeName()),
			},
		}
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		id = "global"
	}

	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), id)},
	}
}

// TODO @shipperizer RoleConverter implementation follows provisional roles API
// GET /roles
// POST /roles
// GET /roles/{id}
// PATCH /roles/{id}
// DELETE /roles/{id}
// GET /roles/{id}/entitlements
// POST /roles/{id}/entitlements
// GET /roles/{id}/entitlements/{e_id} --- not sure we need this?we need to know the type
// DELETE /roles/{id}/entitlements/{e_id}
// POST /roles/{id}/identities/{i_id} --- evaluate if needed, assigning identity to a role
type RoleConverter struct{}

func (c RoleConverter) TypeName() string {
	return ROLE_TYPE
}

func (c RoleConverter) Map(r *http.Request) []Permission {
	role_id := chi.URLParam(r, "id")
	entitlement_id := chi.URLParam(r, "e_id")
	identity_id := chi.URLParam(r, "i_id")

	if entitlement_id != "" && r.Method == http.MethodDelete {
		// DELETE /roles/{id}/entitlements/{e_id} will check for an
		// edit permission on role {id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), role_id)},
		}
	}

	// POST /roles/{id}/entitlements
	if strings.HasSuffix(r.URL.Path, "entitlements") && r.Method == http.MethodPost {
		return []Permission{
			{
				Relation:   CAN_EDIT,
				ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), role_id),
			},
		}
	}

	// TODO @shipperizer this might be canned if we want to stick with PATCH /roles/{id} for user assignment
	if identity_id != "" && r.Method == http.MethodPost {
		// POST /roles/{id}/identities/{i_id} will check for an edit on role {id} and view on {i_id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), role_id)},
			{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, identity_id)},
		}
	}

	if role_id == "" {
		role_id = "global"
	}

	return []Permission{
		{Relation: relation(r), ResourceID: fmt.Sprintf("%s:%s", c.TypeName(), role_id)},
	}
}
