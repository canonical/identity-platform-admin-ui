// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

var (
	identityRules = map[string]string{
		"Credentials": "required",
	}

	// mutually exclusive fields
	credentialsRules = map[string]string{
		"Oidc":     "required_without=Password,excluded_with=Password",
		"Password": "required_without=Oidc,excluded_with=Oidc",
	}
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate
}

func (p *PayloadValidator) setupValidator() {
	p.validator.RegisterStructValidationMapRules(identityRules, CreateIdentityRequest{}.CreateIdentityBody)
	p.validator.RegisterStructValidationMapRules(credentialsRules, CreateIdentityRequest{}.CreateIdentityBody.Credentials)

	p.validator.RegisterStructValidationMapRules(identityRules, UpdateIdentityRequest{}.UpdateIdentityBody)
	p.validator.RegisterStructValidationMapRules(credentialsRules, UpdateIdentityRequest{}.UpdateIdentityBody.Credentials)
}

func (p *PayloadValidator) NeedsValidation(req *http.Request) bool {
	return req.Method == http.MethodPost || req.Method == http.MethodPut
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateIdentity(method, endpoint) {
		createIdentity := new(CreateIdentityRequest)
		if err := json.Unmarshal(body, createIdentity); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(createIdentity)
		validated = true

	} else if p.isUpdateIdentity(method, endpoint) {
		updateIdentity := new(UpdateIdentityRequest)
		if err := json.Unmarshal(body, updateIdentity); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(updateIdentity)
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

func (p *PayloadValidator) isCreateIdentity(method, endpoint string) bool {
	return endpoint == "" && method == http.MethodPost
}

func (p *PayloadValidator) isUpdateIdentity(method, endpoint string) bool {
	return strings.HasPrefix(endpoint, "/") && method == http.MethodPut
}

func NewIdentitiesPayloadValidator(apiKey string) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
