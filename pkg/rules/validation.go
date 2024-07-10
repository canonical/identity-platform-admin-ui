// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	oathkeeper "github.com/ory/oathkeeper-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

var (
	ruleRules = map[string]string{
		// validate slice is not empty
		"Authenticators": "required,min=1,dive,required",
		"Authorizer":     "required",
		"Match":          "required",
		// if not empy, validate every item is not nil and not empty
		"Mutators": "omitempty,dive,required",
		"Upstream": "omitempty,required",
	}

	ruleMatchRules = map[string]string{
		// validate slice is not empty, and each item is a valid uppercase http method
		"Methods": "required,min=1,dive,httpmethod",
		"Url":     "required",
	}

	ruleHandlerRules = map[string]string{
		"Handler": "required",
	}
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate

	logger logging.LoggerInterface
}

func (p *PayloadValidator) setupValidator() {
	p.validator.RegisterAlias("httpmethod", "oneof=GET HEAD POST PUT PATCH DELETE CONNECT OPTIONS TRACE")

	p.validator.RegisterStructValidationMapRules(ruleRules, oathkeeper.Rule{})
	p.validator.RegisterStructValidationMapRules(ruleMatchRules, oathkeeper.RuleMatch{})
	p.validator.RegisterStructValidationMapRules(ruleHandlerRules, oathkeeper.RuleHandler{})
}

func (p *PayloadValidator) NeedsValidation(req *http.Request) bool {
	return req.Method == http.MethodPost || req.Method == http.MethodPut
}

func (p *PayloadValidator) Validate(ctx context.Context, method, endpoint string, body []byte) (context.Context, validator.ValidationErrors, error) {
	validated := false
	var err error

	if p.isCreateOrUpdateRule(method, endpoint) {
		ruleRequest := new(oathkeeper.Rule)
		if err := json.Unmarshal(body, ruleRequest); err != nil {
			p.logger.Error("Json parsing error: ", err)
			return ctx, nil, fmt.Errorf("failed to parse JSON body")
		}

		err = p.validator.Struct(ruleRequest)
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

func (p *PayloadValidator) isCreateOrUpdateRule(method string, endpoint string) bool {
	return (method == http.MethodPost && endpoint == "") || (method == http.MethodPut && strings.HasPrefix(endpoint, "/"))
}

func NewRulesPayloadValidator(apiKey string, logger logging.LoggerInterface) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.logger = logger
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
