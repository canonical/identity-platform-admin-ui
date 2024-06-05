package ui

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/go-chi/chi/v5"
)

const UI = "/ui"

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
	mux.Get(fmt.Sprintf("%s/*", UI), a.uiFiles)
}

// TODO: Validate response when server error handling is implemented
func (a *API) uiFiles(w http.ResponseWriter, r *http.Request) {
	// If the path is not /ui{/,} add the html suffix to make file system serving work
	if ext := path.Ext(r.URL.Path); ext == "" && (r.URL.Path != UI && r.URL.Path != UI+"/") {
		r.URL.Path = fmt.Sprintf("%s.html", r.URL.Path)
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, UI)

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
