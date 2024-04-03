// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"

	"github.com/go-chi/chi/v5"
)

const (
	ROLE_TOKEN_KEY  = "roles"
	GROUP_TOKEN_KEY = "groups"
)

type UpdateRolesRequest struct {
	Roles []string `json:"roles" validate:"required"`
}

type Permission struct {
	Relation string `json:"relation" validate:"required"`
	Object   string `json:"object" validate:"required"`
}

type UpdatePermissionsRequest struct {
	Permissions []Permission `json:"permissions" validate:"required"`
}

type GroupRequest struct {
	ID string `json:"id" validate:"required"`
}

type UpdateIdentitiesRequest struct {
	Identities []string `json:"identities" validate:"required"`
}

// API is the core HTTP object that implements all the HTTP and business logic for the groups
// HTTP API functionality
type API struct {
	service   ServiceInterface
	validator *validator.Validate

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

// RegisterEndpoints hooks up all the endpoints to the server mux passed via the arg
func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/groups", a.handleList)
	mux.Get("/api/v0/groups/{id}", a.handleDetail)
	mux.Post("/api/v0/groups", a.handleCreate)
	mux.Patch("/api/v0/groups/{id}", a.handleUpdate)
	mux.Delete("/api/v0/groups/{id}", a.handleRemove)
	mux.Get("/api/v0/groups/{id}/roles", a.handleListRoles)
	mux.Post("/api/v0/groups/{id}/roles", a.handleAssignRoles)
	mux.Delete("/api/v0/groups/{id}/roles/{r_id}", a.handleRemoveRole)
	mux.Get("/api/v0/groups/{id}/entitlements", a.handleListPermission)
	mux.Patch("/api/v0/groups/{id}/entitlements", a.handleAssignPermission)
	mux.Delete("/api/v0/groups/{id}/entitlements/{e_id}", a.handleRemovePermission)
	mux.Get("/api/v0/groups/{id}/identities", a.handleListIdentities)
	mux.Patch("/api/v0/groups/{id}/identities", a.handleAssignIdentities)
	mux.Delete("/api/v0/groups/{id}/identities/{i_id}", a.handleRemoveIdentities)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterValidatingFunc("groups", a.validatingFunc)
	if err != nil {
		a.logger.Fatal("unexpected validatingFunc already registered for groups")
	}
}

func (a *API) validatingFunc(r *http.Request) validator.ValidationErrors {
	return nil
}

func (a *API) userFromContext(ctx context.Context) *authorization.User {
	// TODO @shipperizer implement the FromContext and NewContext in authorization package
	// see snippet below copied from https://pkg.go.dev/context#Context
	// NewContext returns a new Context that carries value u.
	// func NewContext(ctx context.Context, u *User) context.Context {
	//     return context.WithValue(ctx, userKey, u)
	// }

	// // FromContext returns the User value stored in ctx, if any.
	// func FromContext(ctx context.Context) (*User, bool) {
	//     u, ok := ctx.Value(userKey).(*User)
	//     return u, ok
	// }
	if user := ctx.Value(authorization.USER_CTX); user != nil {
		return user.(*authorization.User)
	}

	user := new(authorization.User)
	user.ID = "anonymous"

	return user
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := a.userFromContext(r.Context())

	groups, err := a.service.ListGroups(
		r.Context(),
		user.ID,
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
	user := a.userFromContext(r.Context())
	role, err := a.service.GetGroup(r.Context(), user.ID, ID)

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
			Data:    []string{role},
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

	group := new(GroupRequest)
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
	user := a.userFromContext(r.Context())
	err = a.service.CreateGroup(r.Context(), user.ID, group.ID)

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
			Message: fmt.Sprintf("Created group %s", group.ID),
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
	permissionUrn := authorization.NewUrnFromURLParam(chi.URLParam(r, "e_id"))

	if permissionUrn == nil {
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
		Permission{Relation: permissionUrn.Relation(), Object: permissionUrn.Object()},
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
			Message: fmt.Sprintf("Removed permission %s for group %s", permissionUrn.ID(), ID),
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleAssignRoles(w http.ResponseWriter, r *http.Request) {
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

	paginator := types.NewTokenPaginator(a.tracer, a.logger)

	if err := paginator.LoadFromRequest(r.Context(), r); err != nil {
		a.logger.Error(err)
	}

	identities, pageToken, err := a.service.ListIdentities(
		r.Context(),
		ID,
		paginator.GetToken(r.Context(), GROUP_TOKEN_KEY),
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

	paginator.SetToken(r.Context(), GROUP_TOKEN_KEY, pageToken)

	pageHeader, err := paginator.PaginationHeader(r.Context())

	if err != nil {
		a.logger.Errorf("error producing pagination header: %s", err)
		pageHeader = ""
	}

	w.Header().Add(types.PAGINATION_HEADER, pageHeader)
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

	a.service = service
	a.validator = validator.New(validator.WithRequiredStructEnabled())
	a.logger = logger
	a.tracer = tracer
	a.monitor = monitor

	return a
}
