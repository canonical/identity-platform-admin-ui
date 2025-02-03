// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"

	"github.com/go-chi/chi/v5"
)

const (
	ROLE_TOKEN_KEY  = "roles"
	GROUP_TOKEN_KEY = "groups"
)

type UpdateRolesRequest struct {
	// validate slice is not nil, and each item is not nil
	Roles []string `json:"roles" validate:"required,dive,required"`
}

type Permission struct {
	Relation string `json:"relation" validate:"required"`
	Object   string `json:"object" validate:"required"`
}

type UpdatePermissionsRequest struct {
	// validate slice is not nil, and each item is not nil
	Permissions []Permission `json:"permissions" validate:"required,dive,required"`
}

type Group struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty" validate:"required,notblank"`
}

type UpdateIdentitiesRequest struct {
	Identities []string `json:"identities" validate:"required,dive,required"`
}

// API is the core HTTP object that implements all the HTTP and business logic for the groups
// HTTP API functionality
type API struct {
	apiKey           string
	service          ServiceInterface
	payloadValidator validation.PayloadValidatorInterface

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

// RegisterEndpoints hooks up all the endpoints to the server mux passed via the arg
func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/groups", a.handleList)
	mux.Get("/api/v0/groups/{id:.+}", a.handleDetail)
	mux.Post("/api/v0/groups", a.handleCreate)
	mux.Patch("/api/v0/groups/{id:.+}", a.handleUpdate)
	mux.Delete("/api/v0/groups/{id:.+}", a.handleRemove)
	mux.Get("/api/v0/groups/{id:.+}/roles", a.handleListRoles)
	mux.Post("/api/v0/groups/{id:.+}/roles", a.handleAssignRoles)
	mux.Delete("/api/v0/groups/{id:.+}/roles/{r_id:.+}", a.handleRemoveRole)
	mux.Get("/api/v0/groups/{id:.+}/entitlements", a.handleListPermission)
	mux.Patch("/api/v0/groups/{id:.+}/entitlements", a.handleAssignPermission)
	mux.Delete("/api/v0/groups/{id:.+}/entitlements/{e_id:.+}", a.handleRemovePermission)
	mux.Get("/api/v0/groups/{id:.+}/identities", a.handleListIdentities)
	mux.Patch("/api/v0/groups/{id:.+}/identities", a.handleAssignIdentities)
	mux.Delete("/api/v0/groups/{id:.+}/identities/{i_id:.+}", a.handleRemoveIdentities)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterPayloadValidator(a.apiKey, a.payloadValidator)
	if err != nil {
		a.logger.Fatalf("unexpected error while registering PayloadValidator for groups, %s", err)
	}
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	principal := authentication.PrincipalFromContext(r.Context())

	groups, err := a.service.ListGroups(
		r.Context(),
		principal.Identifier(),
	)

	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		types.Response{
			Data:    groups,
			Message: "List of groups",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	principal := authentication.PrincipalFromContext(r.Context())
	group, err := a.service.GetGroup(r.Context(), principal.Identifier(), ID)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	if group == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Group not found",
				Status:  http.StatusNotFound,
			},
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    []Group{*group},
			Message: "Group detail",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing request payload",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	group := new(Group)
	if err := json.Unmarshal(body, group); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	if group.ID != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Group ID field is not allowed to be passed in",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	principal := authentication.PrincipalFromContext(r.Context())
	group, err = a.service.CreateGroup(r.Context(), principal.Identifier(), group.Name)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    []Group{*group},
			Message: fmt.Sprintf("Created group %s", group.Name),
			Status:  http.StatusCreated,
		},
	)
}

// handleUpdate is not implemented by choice, product might decide to do it to enhance
// group metadata, we do not support anything on top of simple ID attribute and this is
// not changeable right now due to coupled implementation with OpenFGA
func (a *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("use POST /api/v0/groups/%s/entitlements to assign permissions", ID),
			Status:  http.StatusNotImplemented,
		},
	)
}

// TODO @shipperizer we need to remove all relationships leading to the group
func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	err := a.service.DeleteGroup(r.Context(), ID)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Deleted group %s", ID),
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleListPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	paginator := types.NewTokenPaginator(a.tracer, a.logger)

	if err := paginator.LoadFromRequest(r.Context(), r); err != nil {
		a.logger.Error(err)
	}

	permissions, pageTokens, err := a.service.ListPermissions(
		r.Context(),
		ID,
		paginator.GetAllTokens(r.Context()),
	)

	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	paginator.SetTokens(r.Context(), pageTokens)

	pageHeader, err := paginator.PaginationHeader(r.Context())

	if err != nil {
		a.logger.Errorf("error producing pagination header: %s", err)
		pageHeader = ""
	}

	w.Header().Add(types.PAGINATION_HEADER, pageHeader)
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		types.Response{
			Data:    permissions,
			Message: "List of entitlements",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleListRoles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	roles, err := a.service.ListRoles(
		r.Context(),
		ID,
	)

	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		types.Response{
			Data:    roles,
			Message: "List of roles",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleAssignPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing request payload",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	// we might want to switch to an UpdatePermissionsRequest with additions and removals
	permissions := new(UpdatePermissionsRequest)
	if err := json.Unmarshal(body, permissions); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	err = a.service.AssignPermissions(r.Context(), ID, permissions.Permissions...)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Updated permissions for group %s", ID),
			Status:  http.StatusCreated,
		},
	)
}

func (a *API) handleRemovePermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	permissionURN := authorization.NewURNFromURLParam(chi.URLParam(r, "e_id"))

	if permissionURN == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing entitlement ID",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	err := a.service.RemovePermissions(
		r.Context(),
		ID,
		Permission{Relation: permissionURN.Relation(), Object: permissionURN.Object()},
	)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Removed permission %s for group %s", permissionURN.ID(), ID),
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleAssignRoles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	principal := authentication.PrincipalFromContext(r.Context())

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing request payload",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	roles := new(UpdateRolesRequest)
	if err := json.Unmarshal(body, roles); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	canAssign, err := a.service.CanAssignRoles(r.Context(), principal.Identifier(), roles.Roles...)
	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(rr)
		return
	}

	if !canAssign {
		rr := types.Response{
			Status:  http.StatusForbidden,
			Message: fmt.Sprintf("user %s is not allowed to assign specified roles", principal.Identifier()),
		}

		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(rr)
		return
	}

	err = a.service.AssignRoles(r.Context(), ID, roles.Roles...)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Updated roles for group %s", ID),
			Status:  http.StatusCreated,
		},
	)
}

func (a *API) handleRemoveRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	roleID := chi.URLParam(r, "r_id")

	err := a.service.RemoveRoles(r.Context(), ID, roleID)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Removed role %s from group %s", roleID, ID),
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleListIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	identities, err := a.service.ListIdentities(r.Context(), ID)

	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		types.Response{
			Data:    identities,
			Message: "List of identities",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleAssignIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	principal := authentication.PrincipalFromContext(r.Context())

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing request payload",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	identities := new(UpdateIdentitiesRequest)
	if err := json.Unmarshal(body, identities); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	canAssign, err := a.service.CanAssignIdentities(r.Context(), principal.Identifier(), identities.Identities...)
	if err != nil {
		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(rr)
		return
	}

	if !canAssign {
		rr := types.Response{
			Status:  http.StatusForbidden,
			Message: fmt.Sprintf("user %s is not allowed to assign specified identities", principal.Identifier()),
		}

		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(rr)
		return
	}

	err = a.service.AssignIdentities(r.Context(), ID, identities.Identities...)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Updated identities for group %s", ID),
			Status:  http.StatusCreated,
		},
	)
}

func (a *API) handleRemoveIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	identityID := chi.URLParam(r, "i_id")

	err := a.service.RemoveIdentities(
		r.Context(),
		ID,
		identityID,
	)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("Removed identity %s for group %s", identityID, ID),
			Status:  http.StatusOK,
		},
	)
}

// NewAPI returns an API object responsible for all the roles HTTP handlers
func NewAPI(service ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.apiKey = "groups"
	a.service = service
	a.payloadValidator = NewGroupsPayloadValidator(a.apiKey, logger, tracer)
	a.logger = logger
	a.tracer = tracer
	a.monitor = monitor

	return a
}
