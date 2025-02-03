// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

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
	ROLE_TOKEN_KEY = "roles"
)

type Permission struct {
	Relation string `json:"relation" validate:"required"`
	Object   string `json:"object" validate:"required"`
}

type UpdatePermissionsRequest struct {
	// validate slice is not nil, and each item is not nil
	Permissions []Permission `json:"permissions" validate:"required,dive,required"`
}

type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty" validate:"required,notblank"`
}

// API is the core HTTP object that implements all the HTTP and business logic for the roles
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
	mux.Get("/api/v0/roles", a.handleList)
	mux.Get("/api/v0/roles/{id:.+}", a.handleDetail)
	mux.Post("/api/v0/roles", a.handleCreate)
	mux.Patch("/api/v0/roles/{id:.+}", a.handleUpdate)
	mux.Delete("/api/v0/roles/{id:.+}", a.handleRemove)
	mux.Get("/api/v0/roles/{id:.+}/entitlements", a.handleListPermission)
	mux.Patch("/api/v0/roles/{id:.+}/entitlements", a.handleAssignPermission) // this can only work for assignment unless payload includes add and remove
	mux.Delete("/api/v0/roles/{id:.+}/entitlements/{e_id:.+}", a.handleRemovePermission)
	mux.Get("/api/v0/roles/{id:.+}/groups", a.handleListRoleGroup)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterPayloadValidator("roles", a.payloadValidator)
	if err != nil {
		a.logger.Fatalf("unexpected error while registering PayloadValidator for roles, %s", err)
	}
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	principal := authentication.PrincipalFromContext(r.Context())

	roles, err := a.service.ListRoles(
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
			Data:    roles,
			Message: "List of roles",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")
	principal := authentication.PrincipalFromContext(r.Context())
	role, err := a.service.GetRole(r.Context(), principal.Identifier(), ID)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(rr)

		return
	}

	if role == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Role not found",
				Status:  http.StatusNotFound,
			},
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    []Role{*role},
			Message: "Rule detail",
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

	role := new(Role)
	if err := json.Unmarshal(body, role); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	if role.ID != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Role ID field is not allowed to be passed in",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	principal := authentication.PrincipalFromContext(r.Context())
	role, err = a.service.CreateRole(r.Context(), principal.Identifier(), role.Name)

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
			Data:    []Role{*role},
			Message: fmt.Sprintf("Created role %s", role.Name),
			Status:  http.StatusCreated,
		},
	)
}

// handleUpdate is not implemented by choice, product might decide to do it to enhcance
// role metadata, we do not support anything on top of simple ID attribute and this is
// not changeable right now due to coupled implementation with OpenFGA
func (a *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(
		types.Response{
			Message: fmt.Sprintf("use /api/v0/roles/%s/entitlements to assign permissions", ID),
			Status:  http.StatusNotImplemented,
		},
	)
}

// TODO @shipperizer we need to remove all relationships leading to the role
func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	err := a.service.DeleteRole(r.Context(), ID)

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
			Message: fmt.Sprintf("Deleted role %s", ID),
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

func (a *API) handleListRoleGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ID := chi.URLParam(r, "id")

	roles, err := a.service.ListRoleGroups(r.Context(), ID)

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
			Message: "List of groups",
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
			Message: fmt.Sprintf("Updated permissions for role %s", ID),
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
			Message: fmt.Sprintf("Removed permission %s for role %s", permissionURN.ID(), ID),
			Status:  http.StatusOK,
		},
	)
}

// NewAPI returns an API object responsible for all the roles HTTP handlers
func NewAPI(service ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	a.payloadValidator = NewRolesPayloadValidator(a.apiKey, tracer, monitor, logger)
	a.logger = logger
	a.tracer = tracer
	a.monitor = monitor

	return a
}
