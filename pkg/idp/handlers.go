// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

const okValue = "ok"

type API struct {
	apiKey           string
	service          ServiceInterface
	payloadValidator validation.PayloadValidatorInterface

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/idps", a.handleList)
	mux.Get("/api/v0/idps/{id:.+}", a.handleDetail)
	mux.Post("/api/v0/idps", a.handleCreate)
	mux.Patch("/api/v0/idps/{id:.+}", a.handlePartialUpdate)
	mux.Delete("/api/v0/idps/{id:.+}", a.handleRemove)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterPayloadValidator(a.apiKey, a.payloadValidator)
	if err != nil {
		a.logger.Fatalf("unexpected error while registering PayloadValidator for idps, %s", err)
	}
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idps, err := a.service.ListResources(r.Context())

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
			Data:    idps,
			Message: "List of IDPs",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ID := chi.URLParam(r, "id")

	idps, err := a.service.GetResource(r.Context(), ID)

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
			Data:    idps,
			Message: "Detail of IDPs",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handlePartialUpdate(w http.ResponseWriter, r *http.Request) {
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

	idp := new(Configuration)
	if err := json.Unmarshal(body, idp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	idps, err := a.service.EditResource(r.Context(), ID, idp)

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
			Data:    idps,
			Message: "Updated IDP",
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

	idp := new(Configuration)
	if err := json.Unmarshal(body, idp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	if idp.ID != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "IDP ID field is not allowed to be passed in",
				Status:  http.StatusBadRequest,
			},
		)

		return
	}

	idps, err := a.service.CreateResource(r.Context(), idp)

	if err != nil {
		status := http.StatusInternalServerError
		rr := types.Response{
			Status:  status,
			Message: err.Error(),
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		types.Response{
			Data:    idps,
			Message: "Created IDP",
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	err := a.service.DeleteResource(r.Context(), id)

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
			Data:    nil,
			Message: "IDP deleted",
			Status:  http.StatusOK,
		},
	)
}

func NewAPI(service ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "idps"

	a.payloadValidator = NewIdPPayloadValidator(a.apiKey, logger)
	a.service = service

	a.logger = logger
	a.tracer = tracer
	a.monitor = monitor

	return a
}
