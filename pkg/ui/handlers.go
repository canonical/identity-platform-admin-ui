// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package ui

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const (
	UIPrefix      = "/ui"
	indexTemplate = "index.html"
)

type Config struct {
	DistFS      fs.FS
	ContextPath string
}

type API struct {
	contextPath string
	fileServer  http.Handler
	distFS      fs.FS

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get(UIPrefix, func(w http.ResponseWriter, r *http.Request) {
		path, err := url.JoinPath("/", a.contextPath, UIPrefix, "/")
		if err != nil {
			a.logger.Error("Failed to construct path: ", err)
			a.internalServerErrorResponse(w, err)
			return
		}
		http.Redirect(w, r, path, http.StatusMovedPermanently)
	})
	mux.Get(UIPrefix+"/", a.uiFiles)
	mux.Get(UIPrefix+"/*", a.uiFiles)
}

func (a *API) internalServerErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(
		types.Response{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		},
	)
}

func (a *API) uiFiles(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, UIPrefix)
	// This is a SPA, every HTML page serves the same `index.html`
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

	// return html with processed template
	if r.URL.Path == "/" {
		t, err := template.New(indexTemplate).ParseFS(a.distFS, indexTemplate)
		if err != nil {
			a.logger.Error("Failed to load %s template: ", indexTemplate, err)
			a.internalServerErrorResponse(w, err)
			return
		}

		// disable cache only for index.html response, with no issues if we return an error
		// `no-store`: This will tell any cache system not to cache the index.html file
		// `no-cache`: This will tell any cache system to check if there is a newer version in the server
		// `must-revalidate`: This will tell any cache system to check for newer version of the file
		// this is considered best practice with SPAs
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)

		normContextPath := a.contextPath
		if !strings.HasSuffix(normContextPath, "/") {
			normContextPath += "/"
		}

		err = t.Execute(w, normContextPath)
		if err != nil {
			a.logger.Error("Failed to process %s template: ", indexTemplate, err)
			a.internalServerErrorResponse(w, err)
			return
		}

		return
	}

	// return requested assets
	a.fileServer.ServeHTTP(w, r)
}

func NewAPI(config *Config, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *API {
	a := new(API)

	a.distFS = config.DistFS
	a.fileServer = http.FileServer(http.FS(config.DistFS))
	a.contextPath = config.ContextPath

	a.tracer = tracer
	a.monitor = monitor
	a.logger = logger

	return a
}
