// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const apiVersion = "v0"

type ValidationRegistryInterface interface {
	ValidationMiddleware(next http.Handler) http.Handler
	RegisterPayloadValidator(key string, vf PayloadValidatorInterface) error
}

type PayloadValidatorInterface interface {
	NeedsValidation(r *http.Request) bool
	Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error)
}

type ValidationRegistry struct {
	validators map[string]PayloadValidatorInterface

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

		payloadValidator, ok := v.validators[key]
		if !ok || !payloadValidator.NeedsValidation(r) {
			next.ServeHTTP(w, r)
			return
		}

		reqBody := r.Body
		defer reqBody.Close()
		body, err := io.ReadAll(reqBody)

		if err != nil {
			badRequestFromError(w, NoBodyError)
			return
		}

		// don't break existing handlers, replace the body that was consumed
		r.Body = io.NopCloser(bytes.NewReader(body))

		endpoint, _ := ApiEndpoint(r.URL.Path, key)
		var validationErr validator.ValidationErrors

		ctx, validationErr, err = payloadValidator.Validate(r.Context(), r.Method, endpoint, body)

		if err != nil {
			badRequestFromError(w, err)
			return
		}

		// handler validation errors
		if validationErr != nil {
			e := NewValidationError("validation errors", validationErr)

			w.WriteHeader(e.Status)
			_ = json.NewEncoder(w).Encode(e)
			return
		}

		// if no errors, proceed with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func badRequestFromError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(
		types.Response{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		},
	)
	return
}

func (v *ValidationRegistry) RegisterPayloadValidator(key string, payloadValidator PayloadValidatorInterface) error {
	if payloadValidator == nil {
		return fmt.Errorf("payloadValidator can't be null")
	}

	if _, ok := v.validators[key]; ok {
		return fmt.Errorf("key is already registered")
	}

	v.validators[key] = payloadValidator

	return nil
}

func (v *ValidationRegistry) apiKey(endpoint string) string {
	after, found := strings.CutPrefix(endpoint, fmt.Sprintf("/api/%s/", apiVersion))
	if !found {
		return ""
	}
	return strings.SplitN(after, "/", 1)[0]
}

// ApiEndpoint returns the endpoint string stripped from the api and version prefix, and the apikey
// it doesn't strip away trailing slash if there is one
func ApiEndpoint(endpoint, apiKey string) (string, bool) {
	after, found := strings.CutPrefix(endpoint, fmt.Sprintf("/api/%s/", apiVersion))
	if !found {
		return "", false
	}

	return strings.CutPrefix(after, apiKey)
}

func NewRegistry(tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *ValidationRegistry {
	v := new(ValidationRegistry)
	v.validators = make(map[string]PayloadValidatorInterface)

	v.tracer = tracer
	v.monitor = monitor
	v.logger = logger

	return v
}
