// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"

	"github.com/go-chi/chi/v5"
	oathkeeper "github.com/ory/oathkeeper-client-go"
)

const DEFAULT_OFFSET int64 = 0

type API struct {
	apiKey           string
	service          ServiceInterface
	payloadValidator validation.PayloadValidatorInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

type PageToken struct {
	Offset int64 `json:"offset" validate:"required"`
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/rules", a.handleList)
	mux.Get("/api/v0/rules/{id:.+}", a.handleDetail)
	mux.Post("/api/v0/rules", a.handleCreate)
	mux.Put("/api/v0/rules/{id:.+}", a.handleUpdate)
	mux.Delete("/api/v0/rules/{id:.+}", a.handleRemove)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterPayloadValidator(a.apiKey, a.payloadValidator)
	if err != nil {
		a.logger.Fatalf("unexpected error while registering PayloadValidator for rules, %s", err)
	}
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pagination := types.ParsePagination(r.URL.Query())

	offset := a.offsetDecode(pagination.PageToken)

	rules, err := a.service.ListRules(r.Context(), offset, pagination.Size)

	if err != nil {

		rr := types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(rr)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		types.Response{
			Data:    rules,
			Message: "List of rules",
			Status:  http.StatusOK,
			Meta: &types.Pagination{
				NavigationTokens: types.NavigationTokens{
					Next: a.offsetTokenEncode(offset + pagination.Size),
					Prev: a.offsetTokenEncode(offset - pagination.Size),
				},
				Size: pagination.Size,
			},
		},
	)
}

func (a *API) offsetTokenEncode(offset int64) string {
	if offset < DEFAULT_OFFSET {
		return ""
	}

	pt := new(PageToken)
	pt.Offset = offset

	token, err := json.Marshal(pt)
	if err != nil {
		a.logger.Warnf("bad page token encoding, defaulting to an empty one: %s", err)
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(token)
}

func (a *API) offsetDecode(pageToken string) int64 {
	if pageToken == "" {
		return DEFAULT_OFFSET
	}

	pt := new(PageToken)

	rawPt, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		a.logger.Warnf("bad page token encoding, defaulting to an empty one: %s", err)
		return DEFAULT_OFFSET
	}

	if err := json.Unmarshal(rawPt, pt); err != nil {
		a.logger.Warnf("bad page token format, defaulting to an empty one: %s", err)
		return DEFAULT_OFFSET
	}

	if err != nil || pt.Offset < DEFAULT_OFFSET {
		a.logger.Warnf("invalid offset, default to %d %s", DEFAULT_OFFSET, err)
		return DEFAULT_OFFSET
	}

	return pt.Offset
}

func (a *API) handleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ruleID := chi.URLParam(r, "id")

	rule, err := a.service.GetRule(r.Context(), ruleID)

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
			Data:    rule,
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

	rule := new(oathkeeper.Rule)
	if err := json.Unmarshal(body, rule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	err = a.service.CreateRule(r.Context(), *rule)

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
			Message: fmt.Sprintf("Created rule %s", *rule.Id),
			Status:  http.StatusCreated,
		},
	)
}

func (a *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ruleID := chi.URLParam(r, "id")

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

	rule := new(oathkeeper.Rule)
	if err := json.Unmarshal(body, rule); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			types.Response{
				Message: "Error parsing JSON payload",
				Status:  http.StatusBadRequest,
			},
		)

		return

	}

	err = a.service.UpdateRule(r.Context(), ruleID, *rule)

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
			Message: fmt.Sprintf("Updated rule %s", *rule.Id),
			Status:  http.StatusOK,
		},
	)
}

func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ruleID := chi.URLParam(r, "id")

	err := a.service.DeleteRule(r.Context(), ruleID)

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
			Message: fmt.Sprintf("Deleted rule %s", ruleID),
			Status:  http.StatusOK,
		},
	)
}

func NewAPI(service ServiceInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)
	a.apiKey = "rules"
	a.payloadValidator = NewRulesPayloadValidator(a.apiKey, logger)
	a.service = service

	a.tracer = tracer
	a.monitor = monitor
	a.logger = logger

	return a
}
