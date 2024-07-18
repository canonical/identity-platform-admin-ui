// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

const (
	ADMIN_PRIVILEGE = "privileged:superuser"
)

// Middleware is the monitoring middleware object implementing Prometheus monitoring
type Middleware struct {
	auth AuthorizerInterface

	// converters
	IdentityConverter
	ClientConverter
	ProviderConverter
	RuleConverter
	SchemeConverter
	RoleConverter
	GroupConverter

	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (mdw *Middleware) mapper(r *http.Request) []Permission {
	// TODO @shipperizer exploit https://pkg.go.dev/github.com/go-chi/chi/v5#URLParam to fetch
	// resource ids like {id}, {<x>_id}, also parse the path to understand type to check against

	if strings.HasPrefix(r.URL.Path, "/api/v0/identities") {
		return mdw.IdentityConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/clients") {
		return mdw.ClientConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/idps") {
		return mdw.ProviderConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/rules") {
		return mdw.RuleConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/schemas") {
		return mdw.SchemeConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/roles") {
		return mdw.RoleConverter.Map(r)
	}
	if strings.HasPrefix(r.URL.Path, "/api/v0/groups") {
		return mdw.GroupConverter.Map(r)
	}

	return []Permission{}
}

func (mdw *Middleware) check(ctx context.Context, userID string, r *http.Request) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	// TODO @shipperizer implement BatchCheck
	for _, permission := range mdw.mapper(r) {
		authorized, err := mdw.auth.Check(
			ctx, userID, permission.Relation, permission.ResourceID, permission.ContextualTuples...,
		)

		select {
		case <-ctx.Done():
			return false, fmt.Errorf("issues connecting to OpenFGA server")
		default:
			// stop at the first failed check
			if !authorized || err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func (mdw *Middleware) skipRoute(r *http.Request) bool {
	switch r.URL.Path {
	case "/api/v0/status", "/api/v0/version", "/api/v0/metrics":
		return true
	case "/api/v0/auth", "/api/v0/auth/callback":
		return true
	default:
		return false
	}
}

func (mwd *Middleware) error(message string, status int, w http.ResponseWriter) {
	r := types.Response{
		Status:  status,
		Message: message,
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(r)
}

func (mdw *Middleware) Authorize() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				// not all endpoints need to validate authorization
				if mdw.skipRoute(r) {
					next.ServeHTTP(w, r)
					return
				}

				// if we got here then `principal` must be != nil
				principal := authentication.PrincipalFromContext(r.Context())
				if principal == nil {
					// should never happen if authentication is configured correctly
					mdw.logger.Error("principal not available in context, cannot proceed with authorization")
					mdw.error("unable to retrieve authenticated user", http.StatusInternalServerError, w)
					return
				}

				isAdmin, err := mdw.auth.Admin().CheckAdmin(r.Context(), principal.Identifier())
				if err != nil {
					mdw.logger.Errorf("failed %s", err)
					mdw.error("failed connecting with OpenFGA", http.StatusInternalServerError, w)

					return
				}

				ID := fmt.Sprintf("user:%s", principal.Identifier())
				// TODO @shipperizer add context timeout
				authorized, err := mdw.check(r.Context(), ID, r)

				if err != nil {
					mdw.logger.Errorf("failed %s", err)
					mdw.error("failed connecting with OpenFGA", http.StatusInternalServerError, w)

					return
				}

				if !authorized {
					mdw.logger.Debugf("%s not authorized to perform operation", ID)
					mdw.error("insufficient permissions to execute operation", http.StatusForbidden, w)

					return
				}

				// TOOD @shipperizer evenutally we will want to add the contextual tuple in the context
				// so it can be used in subsequent calls
				ctx := IsAdminContext(r.Context(), isAdmin)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// NewMiddleware returns a Middleware based on the type of monitor
func NewMiddleware(auth AuthorizerInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Middleware {
	mdw := new(Middleware)

	mdw.auth = auth

	mdw.monitor = monitor
	mdw.logger = logger

	return mdw
}
