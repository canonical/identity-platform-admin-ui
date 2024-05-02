// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"

	"github.com/canonical/identity-platform-admin-ui/internal/validation"
)

type PayloadValidator struct {
	apiKey    string
	validator *validator.Validate
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

	if p.isCreateGroup(method, endpoint) {
		group := new(Group)
		if err := json.Unmarshal(body, group); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(group)
		validated = true
	}

	if p.isUpdateGroup(method, endpoint) {
		// TODO: @barco to implement when the UpdateGroup is implemented
		validated = true
	}

	if p.isAssignRoles(method, endpoint) {
		updateRoles := new(UpdateRolesRequest)
		if err := json.Unmarshal(body, updateRoles); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(updateRoles)
		validated = true
	}

	if p.isAssignPermissions(method, endpoint) {
		updatePermissions := new(UpdatePermissionsRequest)
		if err := json.Unmarshal(body, updatePermissions); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(updatePermissions)
		validated = true
	}

	if p.isAssignIdentities(method, endpoint) {
		updateIdentities := new(UpdateIdentitiesRequest)
		if err := json.Unmarshal(body, updateIdentities); err != nil {
			return ctx, nil, err
		}

		err = p.validator.Struct(updateIdentities)
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

func (p *PayloadValidator) isCreateGroup(method, endpoint string) bool {
	return method == http.MethodPost && endpoint == ""
}

func (p *PayloadValidator) isUpdateGroup(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasPrefix(endpoint, "/")
}

func (p *PayloadValidator) isAssignRoles(method, endpoint string) bool {
	return method == http.MethodPost && strings.HasSuffix(endpoint, "/roles")
}

func (p *PayloadValidator) isAssignPermissions(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasSuffix(endpoint, "/entitlements")
}

func (p *PayloadValidator) isAssignIdentities(method, endpoint string) bool {
	return method == http.MethodPatch && strings.HasSuffix(endpoint, "/identities")
}

func NewGroupsPayloadValidator(apiKey string) *PayloadValidator {
	p := new(PayloadValidator)
	p.apiKey = apiKey
	p.validator = validation.NewValidator()

	p.setupValidator()

	return p
}
