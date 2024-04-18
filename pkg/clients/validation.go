// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	client "github.com/ory/hydra-client-go/v2"

	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

var (
	oauth2ClientRules = map[string]string{
		// if not empy, validate every item is not nil and not empty
		"AllowedCorsOrigins": "omitempty,dive,required",
		"Audience":           "omitempty,dive,required",
		"GrantTypes":         "omitempty,dive,required",
		"ClientName":         "required",
		// if not empty, validate value is one of 'pairwise' and 'public'
		"SubjectType": "omitempty,oneof=pairwise public",
		// if not empty, validate value is one of 'client_secret_basic', 'client_secret_post', 'private_key_jwt' and 'none'
		"TokenEndpointAuthMethod": "omitempty,oneof=client_secret_basic client_secret_post private_key_jwt none",
	}
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate
}

func (p *PayloadValidator) setupValidator() {
	p.validator.RegisterStructValidationMapRules(oauth2ClientRules, client.OAuth2Client{})
}

func (p *PayloadValidator) NeedsValidation(req *http.Request) bool {
	return req.Method == http.MethodPost || req.Method == http.MethodPut
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateClient(method, endpoint) || p.isUpdateClient(method, endpoint) {
		clientRequest := new(client.OAuth2Client)
		if err := json.Unmarshal(body, clientRequest); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(clientRequest)
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

func (p *PayloadValidator) isCreateClient(method string, endpoint string) bool {
	return method == http.MethodPost && endpoint == ""
}

func (p *PayloadValidator) isUpdateClient(method string, endpoint string) bool {
	return method == http.MethodPut && strings.HasPrefix(endpoint, "/")
}

func NewClientsPayloadValidator(apiKey string) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
