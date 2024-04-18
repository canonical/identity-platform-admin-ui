// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	kClient "github.com/ory/kratos-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

var (
	identitySchemaContainerRules = map[string]string{
		"Schema": "required",
	}
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate
}

func (p *PayloadValidator) setupValidator() {
	p.validator.RegisterStructValidationMapRules(
		identitySchemaContainerRules,
		kClient.IdentitySchemaContainer{},
	)
}

func (p *PayloadValidator) NeedsValidation(r *http.Request) bool {
	return r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch
}

func (p *PayloadValidator) isCreateSchema(method, endpoint string) bool {
	return endpoint == "" && method == http.MethodPost
}

func (p *PayloadValidator) isPartialUpdate(method, endpoint string) bool {
	return strings.HasPrefix(endpoint, "/") && method == http.MethodPatch
}

func (p *PayloadValidator) isUpdateDefaultSchema(method, endpoint string) bool {
	return endpoint == "/default" && method == http.MethodPut
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateSchema(method, endpoint) || p.isPartialUpdate(method, endpoint) {
		schema := new(kClient.IdentitySchemaContainer)
		if err := json.Unmarshal(body, schema); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(schema)
		validated = true
	}

	if p.isUpdateDefaultSchema(method, endpoint) {
		schema := new(DefaultSchema)
		if err := json.Unmarshal(body, schema); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(schema)
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

func NewSchemasPayloadValidator(apiKey string) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
