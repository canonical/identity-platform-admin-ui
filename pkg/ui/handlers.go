// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package ui

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/go-chi/chi/v5"
)

type API struct {
	fileServer http.Handler

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

type Config struct {
	DistFS fs.FS
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.NotFound(a.uiFiles)
}

func (a *API) uiFiles(w http.ResponseWriter, r *http.Request) {
	// If a `/api` request ends up here, it means that the route does not exist
	// Do not rely on the UI to return the 404 response
	if strings.HasPrefix(r.URL.Path, "/api") {
		http.NotFound(w, r)
		return
	}
	// This is a SPA, everything HTML page uses the same `index.html`
	if !strings.HasPrefix(r.URL.Path, "/assets") {
		r.URL.Path = "/"
	}

	// Set the UI headers
	// Disables the FLoC (Federated Learning of Cohorts) feature on the browser,
	// preventing the current page from being included in the user's FLoC calculation.
	// FLoC is a proposed replacement for third-party cookies to enable interest-based advertising.
	w.Header().Set("Permissions-Policy", "interest-cohort=()")
	// Prevents the browser from trying to guess the MIME type, which can have security implications.
	// This tells the browser to strictly follow the MIME type provided in the Content-Type header.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Restricts the page from being displayed in a frame, iframe, or object to avoid click jacking attacks,
	// but allows it if the site is navigating to the same origin.
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	// Sets the Content Security Policy (CSP) for the page, which helps mitigate XSS attacks and data injection attacks.
	// The policy allows loading resources (scripts, styles, images, etc.) only from the same origin ('self'), data URLs, and all subdomains of ubuntu.com.
	w.Header().Set("Content-Security-Policy", "default-src 'self' data: https://*.ubuntu.com; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

	a.fileServer.ServeHTTP(w, r)
}

func NewAPI(config *Config, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.fileServer = http.FileServer(http.FS(config.DistFS))

	a.tracer = tracer
	a.monitor = monitor
	a.logger = logger

	return a
}
