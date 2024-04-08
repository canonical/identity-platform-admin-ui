// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"

	"github.com/go-chi/chi/v5"
	oathkeeper "github.com/ory/oathkeeper-client-go"
)

type PageToken struct {
	Offset string `json:"offset" validate:"required"`
}

type API struct {
	service   ServiceInterface
	validator *validator.Validate

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/rules", a.handleList)
	mux.Get("/api/v0/rules/{id}", a.handleDetail)
	mux.Post("/api/v0/rules", a.handleCreate)
	mux.Put("/api/v0/rules/{id}", a.handleUpdate)
	mux.Delete("/api/v0/rules/{id}", a.handleRemove)
}

func (a *API) RegisterValidation(v validation.ValidationRegistryInterface) {
	err := v.RegisterValidatingFunc("rules", a.validatingFunc)
	if err != nil {
		a.logger.Fatal("unexpected validatingFunc already registered for rules")
	}
}

func (a *API) validatingFunc(r *http.Request) validator.ValidationErrors {
	return nil
}

func (a *API) pageDecode(pageToken string, size int) int {
	if pageToken == "" {
		return 0
	}

	pt := new(PageToken)

	rawPt, err := base64.StdEncoding.DecodeString(pageToken)
	if err != nil {
		a.logger.Warnf("bad page token encoding, defaulting to an empty one: %s", err)

		return 0
	}

	if err := json.Unmarshal(rawPt, pt); err != nil {
		a.logger.Warnf("bad page token format, defaulting to an empty one: %s", err)

		return 0
	}

	if err := a.validator.Struct(c); err != nil {
		a.logger.Warnf("invalid page token, defaulting ot an empty one: %s", err)

		return 0
	}

	offset, err := strconv.Atoi(pt.Offset)

	if err != nil {
		a.logger.Warnf("invalid offset, default to 0: %s", err)

		return 0
	}

	return offset / size
}

func (a *API) pageTokenEncode(page, size int) string {
	pt := new(PageToken)
	pt.Offset = fmt.Sprintf("%v", page*size)

	token, err := json.Marshal(pt)
	if err != nil {
		a.logger.Warnf("bad page token encoding, defaulting to an empty one: %s", err)

		return ""
	}

	return base64.StdEncoding.EncodeToString(token)
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pagination := types.ParsePagination(r.URL.Query())

	page := a.pageDecode(pagination.PageToken, int(pagination.Size))

	rules, err := a.service.ListRules(r.Context(), int64(page), pagination.Size)

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
			Data:    rules,
			Message: "List of rules",
			Status:  http.StatusOK,
			Meta: &types.Pagination{
				Next: a.pageTokenEncode(page+1, int(pagination.Size)),
				Size: pagination.Size,
			},
		},
	)
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

func NewAPI(service ServiceInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	a.validator = validator.New(validator.WithRequiredStructEnabled())
	a.logger = logger

	return a
}
