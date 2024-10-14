// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package authorization

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
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

	SYSTEM_OBJECT_PREFIX = "__system__"
	// global is used in place of `*`, reason is OpenFGA has a special meaning for `*`
	// see [public access](https://openfga.dev/docs/modeling/public-access)
	GLOBAL_ACCESS_OBJECT_NAME = SYSTEM_OBJECT_PREFIX + "global"
)

type Permission struct {
	Relation         string
	ResourceID       string
	ContextualTuples []openfga.Tuple
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

func (c IdentityConverter) MapV1(r *http.Request) []Permission {
	// TODO @shipperizer @mrbarco
	// the following endpoints will need permissions inspection on the payload
	// on top of uri permission validation, operator needs to have those permissions
	// Add or remove entitlement to/from an identity. (CAN_VIEW on entitlements)
	// (PATCH /identities/{id}/entitlements)
	// PatchIdentitiesItemEntitlements(w http.ResponseWriter, r *http.Request, id string)
	// Add or remove the identity to/from groups. (CAN_EDIT on groups)
	// (PATCH /identities/{id}/groups)
	// PatchIdentitiesItemGroups(w http.ResponseWriter, r *http.Request, id string)
	// Add or remove the identity to/from roles. (CAN_EDIT on groups)
	// (PATCH /identities/{id}/roles)
	// PatchIdentitiesItemRoles(w http.ResponseWriter, r *http.Request, id string)

	return c.MapV0(r)
}

func (c IdentityConverter) MapV0(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")
	var resourceId string
	var contextualTuples []openfga.Tuple

	if id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = append(
			contextualTuples,
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
		)
	} else {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), id)
	}

	// Admins have privileged access to resource
	contextualTuples = append(
		contextualTuples,
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	)

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}

type ClientConverter struct{}

func (c ClientConverter) TypeName() string {
	return CLIENT_TYPE
}

func (c ClientConverter) MapV0(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")
	var resourceId string
	var contextualTuples []openfga.Tuple

	if id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = append(
			contextualTuples,
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
		)
	} else {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), id)
	}

	// Admins have privileged access to resource
	contextualTuples = append(
		contextualTuples,
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	)

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}

type ProviderConverter struct{}

func (c ProviderConverter) TypeName() string {
	return PROVIDER_TYPE
}

func (c ProviderConverter) MapV1(r *http.Request) []Permission {
	if strings.HasPrefix(r.URL.Path, "/api/v1/authentication/providers") {
		return []Permission{}
	}

	return c.MapV0(r)
}

func (c ProviderConverter) MapV0(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")
	var resourceId string
	var contextualTuples []openfga.Tuple

	if id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = append(
			contextualTuples,
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
		)
	} else {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), id)
	}

	// Admins have privileged access to resource
	contextualTuples = append(
		contextualTuples,
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	)

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}

type RuleConverter struct{}

func (c RuleConverter) TypeName() string {
	return RULE_TYPE
}

func (c RuleConverter) MapV0(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")
	var resourceId string
	var contextualTuples []openfga.Tuple

	if id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = append(
			contextualTuples,
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
		)
	} else {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), id)
	}

	// Admins have privileged access to resource
	contextualTuples = append(
		contextualTuples,
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	)

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}

type SchemeConverter struct{}

func (c SchemeConverter) TypeName() string {
	return SCHEME_TYPE
}

func (c SchemeConverter) MapV0(r *http.Request) []Permission {
	id := chi.URLParam(r, "id")
	var resourceId string
	var contextualTuples []openfga.Tuple

	if r.URL.Path == "/api/v0/schemas/default" {
		resourceId = fmt.Sprintf("%s:**DEFAULT**", c.TypeName())
	} else if id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = append(
			contextualTuples,
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
		)
	} else {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), id)
	}

	// Admins have privileged access to resource
	contextualTuples = append(
		contextualTuples,
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	)

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
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

func (c RoleConverter) MapV1(r *http.Request) []Permission {
	// TODO @shipperizer @mrbarco
	// the following endpoints will need permissions inspection on the payload
	// on top of uri permission validation, operator needs to have those permissions
	// // Add or remove a direct entitlements to/from a role. (CAN_VIEW on entitlements)
	// // (PATCH /roles/{id}/entitlements)
	// PatchRolesItemEntitlements(w http.ResponseWriter, r *http.Request, id string)
	return c.MapV0(r)
}

func (c RoleConverter) MapV0(r *http.Request) []Permission {
	role_id := chi.URLParam(r, "id")
	entitlement_id := chi.URLParam(r, "e_id")
	identity_id := chi.URLParam(r, "i_id")
	resourceId := fmt.Sprintf("%s:%s", c.TypeName(), role_id)

	var contextualTuples []openfga.Tuple = []openfga.Tuple{
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	}

	if entitlement_id != "" && r.Method == http.MethodDelete {
		// DELETE /roles/{id}/entitlements/{e_id} will check for an
		// edit permission on role {id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: resourceId, ContextualTuples: contextualTuples},
		}
	}

	// POST /roles/{id}/entitlements
	if strings.HasSuffix(r.URL.Path, "entitlements") && r.Method == http.MethodPost {
		return []Permission{
			{
				Relation:         CAN_EDIT,
				ResourceID:       resourceId,
				ContextualTuples: contextualTuples,
			},
		}
	}

	// TODO @shipperizer this might be canned if we want to stick with PATCH /roles/{id} for user assignment
	if identity_id != "" && r.Method == http.MethodPost {
		// POST /roles/{id}/identities/{i_id} will check for an edit on role {id} and view on {i_id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: resourceId, ContextualTuples: contextualTuples},
			{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, identity_id)},
		}
	}

	if role_id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		globalResource := fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = []openfga.Tuple{
			*openfga.NewTuple("user:*", CAN_VIEW, globalResource),
			*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, globalResource),
		}
	}
	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}

// GET /api/v0/groups
// GET /api/v0/groups/{id}
// POST /api/v0/groups
// PATCH /api/v0/groups/{id}
// DELETE /api/v0/groups/{id}
// GET /api/v0/groups/{id}/roles
// POST /api/v0/groups/{id}/roles
// DELETE /api/v0/groups/{id}/roles/{r_id}
// GET /api/v0/groups/{id}/entitlements
// POST /api/v0/groups/{id}/entitlements
// DELETE /api/v0/groups/{id}/entitlements/{e_id}
// GET /api/v0/groups/{id}/identities
// PATCH /api/v0/groups/{id}/identities
// DELETE /api/v0/groups/{id}/identities/{i_id}
type GroupConverter struct{}

func (c GroupConverter) rolesAssignementCheck(r *http.Request) ([]Permission, error) {
	// TODO @shipperizer payloads should be centralized in a separate package to avoid circular
	// imports and duplications
	type UpdateRolesRequest struct {
		// validate slice is not nil, and each item is not nil
		Roles []string `json:"roles" validate:"required,dive,required"`
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	rData := new(UpdateRolesRequest)
	if err := json.Unmarshal(body, rData); err != nil {
		return nil, err
	}

	permissions := make([]Permission, 0)

	for _, role := range rData.Roles {
		permissions = append(permissions, Permission{
			Relation:   CAN_VIEW_RELATION,
			ResourceID: RoleForTuple(role),
		},
		)
	}

	return permissions, nil
}

func (c GroupConverter) identitiesAssignementCheck(r *http.Request) ([]Permission, error) {
	// TODO @shipperizer payloads should be centralized in a separate package to avoid circular
	// imports and duplications
	type UpdateIdentitiesRequest struct {
		Identities []string `json:"identities" validate:"required,dive,required"`
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	rData := new(UpdateIdentitiesRequest)
	if err := json.Unmarshal(body, rData); err != nil {
		return nil, err
	}

	permissions := make([]Permission, 0)

	for _, identity := range rData.Identities {
		permissions = append(permissions, Permission{
			Relation:   CAN_VIEW_RELATION,
			ResourceID: IdentityForTuple(identity),
		},
		)
	}

	return permissions, nil
}

func (c GroupConverter) TypeName() string {
	return GROUP_TYPE
}

func (c GroupConverter) MapV1(r *http.Request) []Permission {
	// TODO @shipperizer @mrbarco
	// the following endpoints will need permissions inspection on the payload
	// on top of uri permission validation, operator needs to have those permissions
	// Add or remove a direct entitlements to/from a group. (CAN_VIEW on entitlements)
	// (PATCH /groups/{id}/entitlements)
	// PatchGroupsItemEntitlements(w http.ResponseWriter, r *http.Request, id string)
	// Add or remove identities to/from a group. (CAN_EDIT on identitites)
	// (PATCH /groups/{id}/identities)
	// PatchGroupsItemIdentities(w http.ResponseWriter, r *http.Request, id string)
	// Add or remove a role to/from the group. (CAN_EDIT on roles)
	// (PATCH /groups/{id}/roles)
	// PatchGroupsItemRoles(w http.ResponseWriter, r *http.Request, id string)
	return c.MapV0(r)
}

func (c GroupConverter) MapV0(r *http.Request) []Permission {
	group_id := chi.URLParam(r, "id")
	role_id := chi.URLParam(r, "r_id")
	identity_id := chi.URLParam(r, "i_id")
	entitlement_id := chi.URLParam(r, "e_id")
	resourceId := fmt.Sprintf("%s:%s", c.TypeName(), group_id)

	var contextualTuples []openfga.Tuple = []openfga.Tuple{
		*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
	}

	// DELETE /api/v0/groups/{id}/entitlements/{e_id}
	if entitlement_id != "" && r.Method == http.MethodDelete {
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: resourceId, ContextualTuples: contextualTuples},
		}
	}

	if identity_id != "" && r.Method == http.MethodDelete {
		// DELETE /groups/{id}/identities/{i_id} will check for an
		// edit permission on group {id} and view permissions on identity {i_id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: resourceId, ContextualTuples: contextualTuples},
			// TODO @shipperizer this checks for an identity being present, even though the relation
			// is between a user type and the group
			// we need to work and sync users and identities or drop the check below as this would
			// be missing 100% of the times
			{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", IDENTITY_TYPE, identity_id)},
		}
	}

	// PATCH /api/v0/groups/{id}/identities
	// TODO @shipperizer payload inspection needs to be dealt with in the handler to make sure
	// identities are viewable
	if strings.HasSuffix(r.URL.Path, "identities") && r.Method == http.MethodPatch {
		// TODO @shipperizer need to break the chain if error is != nil
		ps, _ := c.identitiesAssignementCheck(r)

		ps = append(
			ps,
			Permission{
				Relation:         CAN_EDIT,
				ResourceID:       resourceId,
				ContextualTuples: contextualTuples,
			},
		)
		return ps
	}

	// DELETE /api/v0/groups/{id}/entitlements
	// POST /api/v0/groups/{id}/entitlements
	if strings.HasSuffix(r.URL.Path, "entitlements") && (r.Method == http.MethodDelete || r.Method == http.MethodPost) {
		return []Permission{
			{
				Relation:         CAN_EDIT,
				ResourceID:       resourceId,
				ContextualTuples: contextualTuples,
			},
		}
	}

	// POST /api/v0/groups/{id}/roles
	// TODO @shipperizer payload inspection needs to be dealt with in the handler to make sure
	// roles are viewable
	if strings.HasSuffix(r.URL.Path, "roles") && r.Method == http.MethodPost {
		// TODO @shipperizer need to break the chain if error is != nil
		ps, _ := c.rolesAssignementCheck(r)

		ps = append(
			ps,
			Permission{
				Relation:         CAN_EDIT,
				ResourceID:       resourceId,
				ContextualTuples: contextualTuples,
			},
		)

		return ps
	}

	if role_id != "" && r.Method == http.MethodDelete {
		// DELETE /groups/{id}/roles/{r_id} will check for an
		// edit permission on group {id} and view permissions on role {r_id}
		return []Permission{
			{Relation: CAN_EDIT, ResourceID: resourceId, ContextualTuples: contextualTuples},
			{Relation: CAN_VIEW, ResourceID: fmt.Sprintf("%s:%s", ROLE_TYPE, role_id)},
		}
	}

	if group_id == "" {
		resourceId = fmt.Sprintf("%s:%s", c.TypeName(), GLOBAL_ACCESS_OBJECT_NAME)
		contextualTuples = []openfga.Tuple{
			*openfga.NewTuple("user:*", CAN_VIEW, resourceId),
			*openfga.NewTuple(ADMIN_OBJECT, PRIVILEGED_RELATION, resourceId),
		}
	}

	return []Permission{
		{Relation: relation(r), ResourceID: resourceId, ContextualTuples: contextualTuples},
	}
}
