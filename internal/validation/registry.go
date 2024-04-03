// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const apiVersion = "v0"

type ValidationRegistryInterface interface {
	ValidationMiddleware(next http.Handler) http.Handler
	RegisterValidatingFunc(key string, vf ValidatingFunc) error
}

type ValidatingFunc func(r *http.Request) validator.ValidationErrors

type ValidationRegistry struct {
	validatingFuncs map[string]ValidatingFunc

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (v *ValidationRegistry) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := v.tracer.Start(r.Context(), "validator.ValidationRegistry.ValidationMiddleware")
		defer span.End()

		r = r.WithContext(ctx)
		key := v.apiKey(r.URL.Path)

		vf, ok := v.validatingFuncs[key]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		if validationErr := vf(r); validationErr != nil {
			e := NewValidationError(validationErr)

			w.WriteHeader(e.Status)
			_ = json.NewEncoder(w).Encode(e)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (v *ValidationRegistry) RegisterValidatingFunc(key string, vf ValidatingFunc) error {
	if vf == nil {
		return fmt.Errorf("validatingFunc can't be null")
	}

	if _, ok := v.validatingFuncs[key]; ok {
		return fmt.Errorf("key is already registered")
	}

	v.validatingFuncs[key] = vf

	return nil
}

func (v *ValidationRegistry) apiKey(endpoint string) string {
	after, found := strings.CutPrefix(endpoint, fmt.Sprintf("/api/%s/", apiVersion))
	if !found {
		return ""
	}
	return strings.SplitN(after, "/", 1)[0]
}

func NewRegistry(tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *ValidationRegistry {
	v := new(ValidationRegistry)
	v.validatingFuncs = make(map[string]ValidatingFunc)

	v.tracer = tracer
	v.monitor = monitor
	v.logger = logger

	return v
}
