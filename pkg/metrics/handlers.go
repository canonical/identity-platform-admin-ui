// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
)

type API struct {
	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/api/v0/metrics", a.prometheusHTTP)
}

func (a *API) prometheusHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func (a *API) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.prometheusHTTP(w, r)
	})
}

func NewAPI(logger logging.LoggerInterface) *API {
	a := new(API)

	a.logger = logger

	return a
}
