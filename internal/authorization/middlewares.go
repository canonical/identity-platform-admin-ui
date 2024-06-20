// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package authorization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

const (
	// custom header for the time being
	TOKEN_HEADER                 = "X-Authorization"
	ADMIN_PRIVILEGE              = "privileged:superuser"
	ADMIN_CTX       AdminContext = "adminCtx"
	USER_CTX        UserContext  = "userCtx"
)

type UserContext string
type AdminContext string

type User struct {
	ID string
}

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

func (mdw *Middleware) transformToken(token string) *User {
	user := new(User)
	user.ID = "admin"

	// TODO @shipperizer rudimentary base64 username
	if token != "" {
		ID, _ := base64.StdEncoding.DecodeString(token)
		user.ID = string(ID)
	}

	return user

}

// TODO @shipperizer move this to a separate middleware once implementation of authorization is starting
func (mdw *Middleware) admin(r *http.Request) bool {
	// TODO @shipperizer implement how to fetch user from cookie or header
	user := mdw.transformToken(r.Header.Get(TOKEN_HEADER))

	isAdmin, err := mdw.auth.Check(context.Background(), fmt.Sprintf("user:%s", user.ID), "admin", ADMIN_PRIVILEGE)

	return isAdmin && err == nil
}

// TODO @shipperizer move this to a separate middleware once implementation of authorization is starting
func (mdw *Middleware) user(r *http.Request) *User {
	// TODO @shipperizer implement how to fetch user from cookie or header
	user := mdw.transformToken(r.Header.Get(TOKEN_HEADER))

	return user
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
		authorized, err := mdw.auth.Check(ctx, userID, permission.Relation, permission.ResourceID)

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
	case "/api/v0/status", "/api/v0/version":
		return true
	case "/api/v0/metrics":
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

				user := mdw.user(r)
				isAdmin := mdw.admin(r)

				ID := fmt.Sprintf("user:%s", user.ID)
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

				// pass on context with user object
				ctx := context.WithValue(r.Context(), USER_CTX, user)
				// TOOD @shipperizer evenutally we will want to add the contextual tuple in the context
				// so it can be used in subsequent calls
				ctx = context.WithValue(ctx, ADMIN_CTX, isAdmin)

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
