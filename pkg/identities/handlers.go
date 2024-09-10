// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	kClient "github.com/ory/kratos-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
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
	apiKey           string
	service          ServiceInterface
	payloadValidator validation.PayloadValidatorInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
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
	err := v.RegisterPayloadValidator(a.apiKey, a.payloadValidator)

	if err != nil {
		a.logger.Fatalf("unexpected error while registering PayloadValidator for identities, %s", err)
	}
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data: ids.Identities,
			Meta: &types.Pagination{
				NavigationTokens: types.NavigationTokens{
					Next: ids.Tokens.Next,
					Prev: ids.Tokens.Prev,
				},
				Size: pagination.Size,
			},
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

	createdIdentity := &ids.Identities[0]
	err = a.service.SendUserCreationEmail(r.Context(), createdIdentity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		)

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

func NewAPI(service ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "identities"
	a.service = service

	a.payloadValidator = NewIdentitiesPayloadValidator(a.apiKey, logger)

	a.tracer = tracer
	a.monitor = monitor
	a.logger = logger

	return a
}
