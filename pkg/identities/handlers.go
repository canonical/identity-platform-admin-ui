// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	kClient "github.com/ory/kratos-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

// CreateIdentityRequest is used as a proxy struct
type CreateIdentityRequest struct {
	kClient.CreateIdentityBody
}

// UpdateIdentityRequest is used as a proxy struct
type UpdateIdentityRequest struct {
	kClient.UpdateIdentityBody
}

type API struct {
	service   ServiceInterface
	validator *validator.Validate

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/identities", a.handleList)
	mux.Get("/api/v0/identities/{id:.+}", a.handleDetail)
	mux.Post("/api/v0/identities", a.handleCreate)
	mux.Put("/api/v0/identities/{id:.+}", a.handleUpdate)
	// mux.Patch("/api/v0/identities/{id:.+}", a.handlePartialUpdate)
	mux.Delete("/api/v0/identities/{id:.+}", a.handleRemove)
	// mux.Delete("/api/v0/identities/{id:.+}/sessions", a.handleSessionRemove)
	// mux.Delete("/api/v0/identities/{id:.+}/credentials/{type}", a.handleCrededntialRemove)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterValidatingFunc("identities", a.validatingFunc)
	if err != nil {
		a.logger.Fatal("unexpected validatingFunc already registered for identities")
	}
}

func (a *API) validatingFunc(r *http.Request) validator.ValidationErrors {
	return nil
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pagination := types.ParsePagination(r.URL.Query())

	credID := r.URL.Query().Get("credID")

	ids, err := a.service.ListIdentities(r.Context(), pagination.Size, pagination.PageToken, credID)

	if err != nil {
		rr := a.error(ids.Error)

		w.WriteHeader(rr.Status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	// TODO @shipperizer improve on this, see if better to stick with link headers
	pagination.Next = ids.Tokens.Next
	pagination.Prev = ids.Tokens.Prev

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    ids.Identities,
			Meta:    pagination,
			Message: "List of identities",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	credID := chi.URLParam(r, "id")

	ids, err := a.service.GetIdentity(r.Context(), credID)

	if err != nil {
		rr := a.error(ids.Error)

		w.WriteHeader(rr.Status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    ids.Identities,
			Message: "Identity detail",
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

	identity := new(CreateIdentityRequest)
	if err := json.Unmarshal(body, identity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	ids, err := a.service.CreateIdentity(r.Context(), &identity.CreateIdentityBody)

	if err != nil {
		rr := a.error(ids.Error)

		w.WriteHeader(rr.Status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    ids.Identities,
			Message: "Created identity",
			Status:  http.StatusCreated,
		},
	)
}

func (a *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	credID := chi.URLParam(r, "id")

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

	identity := new(UpdateIdentityRequest)
	if err := json.Unmarshal(body, identity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	ids, err := a.service.UpdateIdentity(r.Context(), credID, &identity.UpdateIdentityBody)

	if err != nil {
		rr := a.error(ids.Error)

		w.WriteHeader(rr.Status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    ids.Identities,
			Message: "Updated identity",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	credID := chi.URLParam(r, "id")

	identities, err := a.service.DeleteIdentity(r.Context(), credID)

	if err != nil {
		rr := a.error(identities.Error)

		w.WriteHeader(rr.Status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    identities.Identities,
			Message: "Identity Deleted",
			Status:  http.StatusOK,
		},
	)
}

// TODO @shipperizer encapsulate kClient.GenericError into a service error to remove library dependency
func (a *API) error(e *kClient.GenericError) types.Response {
	r := types.Response{
		Status: http.StatusInternalServerError,
	}

	if e.Reason != nil {
		r.Message = *e.Reason
	}

	if e.Code != nil {
		r.Status = int(*e.Code)
	}

	return r

}

func NewAPI(service ServiceInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	a.validator = validator.New(validator.WithRequiredStructEnabled())
	a.logger = logger

	return a
}
