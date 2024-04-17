// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

type API struct {
	apiKey           string
	service          ServiceInterface
	payloadValidator validation.PayloadValidatorInterface

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/clients", a.ListClients)
	mux.Post("/api/v0/clients", a.CreateClient)
	mux.Get("/api/v0/clients/{id}", a.GetClient)
	mux.Put("/api/v0/clients/{id}", a.UpdateClient)
	mux.Delete("/api/v0/clients/{id}", a.DeleteClient)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterPayloadValidator(a.apiKey, a.payloadValidator)
	if err != nil {
		a.logger.Fatalf("unexpected validatingFunc already registered for clients, %s", err)
	}
}

func (a *API) WriteJSONResponse(w http.ResponseWriter, data interface{}, msg string, status int, links interface{}, meta *types.Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(
		types.Response{
			Data:    data,
			Meta:    meta,
			Message: msg,
			Status:  status,
		},
	)

	if err != nil {
		a.logger.Errorf("Unexpected error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *API) GetClient(w http.ResponseWriter, r *http.Request) {
	clientId := chi.URLParam(r, "id")

	res, e := a.service.GetClient(r.Context(), clientId)
	if e != nil {
		a.logger.Errorf("Unexpected error: %s", e)
		a.WriteJSONResponse(w, nil, "Unexpected internal error", http.StatusInternalServerError, nil, nil)
		return
	}
	if res.ServiceError != nil {
		a.WriteJSONResponse(w, res.ServiceError, "Failed to get client", res.ServiceError.StatusCode, nil, nil)
		return
	}

	a.WriteJSONResponse(w, res.Resp, "Client info", http.StatusOK, nil, nil)
}

func (a *API) DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientId := chi.URLParam(r, "id")

	res, e := a.service.DeleteClient(r.Context(), clientId)
	if e != nil {
		a.logger.Errorf("Unexpected error: %s", e)
		a.WriteJSONResponse(w, nil, "Unexpected internal error", http.StatusInternalServerError, nil, nil)
		return
	}
	if res.ServiceError != nil {
		a.WriteJSONResponse(w, res.ServiceError, "Failed to delete client", res.ServiceError.StatusCode, nil, nil)
		return
	}

	a.WriteJSONResponse(w, "", "Client deleted", http.StatusOK, nil, nil)
}

func (a *API) CreateClient(w http.ResponseWriter, r *http.Request) {
	// TODO @nsklikas: Limit request params?
	json_data, err := io.ReadAll(r.Body)
	if err != nil {
		a.WriteJSONResponse(w, nil, "Failed to parse request body", http.StatusBadRequest, nil, nil)
		return
	}
	c, err := a.service.UnmarshalClient(json_data)
	if err != nil {
		a.logger.Debugf("Failed to unmarshal JSON: %s", err)
		a.WriteJSONResponse(w, nil, "Failed to parse request body", http.StatusBadRequest, nil, nil)
		return
	}

	res, e := a.service.CreateClient(r.Context(), c)
	if e != nil {
		a.logger.Errorf("Unexpected error: %s", e)
		a.WriteJSONResponse(w, nil, "Unexpected internal error", http.StatusInternalServerError, nil, nil)
		return
	}
	if res.ServiceError != nil {
		a.WriteJSONResponse(w, res.ServiceError, "Failed to create client", res.ServiceError.StatusCode, nil, nil)
		return
	}

	a.WriteJSONResponse(w, res.Resp, "Created client", http.StatusCreated, nil, nil)
}

func (a *API) UpdateClient(w http.ResponseWriter, r *http.Request) {
	clientId := chi.URLParam(r, "id")

	json_data, err := io.ReadAll(r.Body)
	if err != nil {
		a.logger.Debugf("Failed to read response body: %s", err)
		a.WriteJSONResponse(w, nil, "Failed to parse request body", http.StatusBadRequest, nil, nil)
		return
	}
	// TODO @nsklikas: Limit request params?
	c, err := a.service.UnmarshalClient(json_data)
	if err != nil {
		a.logger.Debugf("Failed to unmarshal JSON: %s", err)
		a.WriteJSONResponse(w, nil, "Failed to parse request body", http.StatusBadRequest, nil, nil)
		return
	}
	c.SetClientId(clientId)

	res, e := a.service.UpdateClient(r.Context(), c)
	if e != nil {
		a.logger.Errorf("Unexpected error: %s", e)
		a.WriteJSONResponse(w, nil, "Unexpected internal error", http.StatusInternalServerError, nil, nil)
		return
	}
	if res.ServiceError != nil {
		a.WriteJSONResponse(w, res.ServiceError, "Failed to update client", res.ServiceError.StatusCode, nil, nil)
		return
	}

	a.WriteJSONResponse(w, res.Resp, "Updated client", http.StatusOK, nil, nil)
}

func (a *API) ListClients(w http.ResponseWriter, r *http.Request) {
	req, err := a.parseListClientsRequest(r)
	if err != nil {
		a.WriteJSONResponse(w, nil, "Failed to parse request", http.StatusBadRequest, nil, nil)
		return
	}

	res, e := a.service.ListClients(r.Context(), req)
	if e != nil {
		a.logger.Errorf("Unexpected error: %s", e)
		a.WriteJSONResponse(w, nil, "Unexpected internal error", http.StatusInternalServerError, nil, nil)
		return
	}
	if res.ServiceError != nil {
		a.WriteJSONResponse(w, res.ServiceError, "Failed to fetch clients", res.ServiceError.StatusCode, nil, nil)
		return
	}

	meta := new(types.Pagination)
	meta.NavigationTokens = res.Tokens
	meta.Size = req.Size

	a.WriteJSONResponse(w, res.Resp, "List of clients", http.StatusOK, nil, meta)
}

func (a *API) parseListClientsRequest(r *http.Request) (*ListClientsRequest, error) {
	q := r.URL.Query()

	cn := q.Get("client_name")
	owner := q.Get("owner")
	page_token := q.Get("page_token")
	s := q.Get("size")

	var size int = 200
	if s != "" {
		var err error
		size, err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}
	return NewListClientsRequest(cn, owner, page_token, size), nil
}

func NewAPI(service ServiceInterface, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "clients"

	a.service = service
	//a.payloadValidator = NewClientsPayloadValidator(a.apiKey)
	a.logger = logger

	return a
}
