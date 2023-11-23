package rules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"

	"github.com/go-chi/chi/v5"
	oathkeeper "github.com/ory/oathkeeper-client-go"
)

type API struct {
	service ServiceInterface

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/rules", a.handleList)
	mux.Get("/api/v0/rules/{id}", a.handleDetail)
	mux.Post("/api/v0/rules", a.handleCreate)
	mux.Put("/api/v0/rules/{id}", a.handleUpdate)
	mux.Delete("/api/v0/rules/{id}", a.handleRemove)
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pagination := types.ParsePagination(r.URL.Query())

	rules, err := a.service.ListRules(r.Context(), pagination.Page, pagination.Size)

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

	a.logger = logger

	return a
}
