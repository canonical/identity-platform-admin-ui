// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (p *PayloadValidator) setupValidator() {
	_ = p.validator.RegisterValidation("notblank", validators.NotBlank)
}

func (p *PayloadValidator) NeedsValidation(r *http.Request) bool {
	return r.Method == http.MethodPost || r.Method == http.MethodPatch
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateRole(method, endpoint) {
		roleRequest := new(Role)
		if err := json.Unmarshal(body, roleRequest); err != nil {
			p.logger.Error("Json parsing error: ", err)
			return ctx, nil, fmt.Errorf("failed to parse JSON body")
		}

		err = p.validator.Struct(roleRequest)
		validated = true
	}

	if p.isUpdateRole(method, endpoint) {
		// TODO: @barco to implement when the UpdateGroup is implemented
		validated = true
	}

	if p.isAssignPermissions(method, endpoint) {
		updatePermissions := new(UpdatePermissionsRequest)
		if err := json.Unmarshal(body, updatePermissions); err != nil {
			p.logger.Error("Json parsing error: ", err)
			return ctx, nil, fmt.Errorf("failed to parse JSON body")
		}

		err = p.validator.Struct(updatePermissions)
		validated = true
	}

	if !validated {
		return ctx, nil, validation.NoMatchError(p.apiKey)
	}

	if err == nil {
		return ctx, nil, nil
	}

	return ctx, err.(validator.ValidationErrors), nil
}

func (p *PayloadValidator) isCreateRole(method, endpoint string) bool {
	return method == http.MethodPost && endpoint == ""
}

func (p *PayloadValidator) isUpdateRole(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasPrefix(endpoint, "/")
}

func (p *PayloadValidator) isAssignPermissions(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasSuffix(endpoint, "/entitlements")
}

func NewRolesPayloadValidator(apiKey string, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.validator = validation.NewValidator()

	p.tracer = tracer
	p.monitor = monitor
	p.logger = logger

	p.setupValidator()

	return p
}
