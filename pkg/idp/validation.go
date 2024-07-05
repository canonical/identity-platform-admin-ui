// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate

	logger logging.LoggerInterface
}

func (p *PayloadValidator) setupValidator() {
	// validate Provider to be one of the supported ones
	p.validator.RegisterAlias("supported_provider", "oneof=generic google github githubapp gitlab microsoft discord slack facebook auth0 vk yandex apple spotify netid dingtalk linkedin patreon")
}

func (p *PayloadValidator) NeedsValidation(r *http.Request) bool {
	return r.Method == http.MethodPost || r.Method == http.MethodPatch
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateIdP(method, endpoint) || p.isPartialUpdateIdP(method, endpoint) {
		conf := new(Configuration)
		if err := json.Unmarshal(body, conf); err != nil {
			p.logger.Error("Json parsing error: ", err)
			return ctx, nil, fmt.Errorf("failed to parse JSON body")
		}

		err = p.validator.Struct(conf)
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

func (p *PayloadValidator) isCreateIdP(method, endpoint string) bool {
	return method == http.MethodPost && endpoint == ""
}

func (p *PayloadValidator) isPartialUpdateIdP(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasPrefix(endpoint, "/")
}

func NewIdPPayloadValidator(apiKey string, logger logging.LoggerInterface) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.logger = logger
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
