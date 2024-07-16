// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"context"

	oathkeeper "github.com/ory/oathkeeper-client-go"
)

type AuthorizerInterface interface {
	SetCreateRuleEntitlements(context.Context, string) error
	SetDeleteRuleEntitlements(context.Context, string) error
}

type ServiceInterface interface {
	ListRules(context.Context, int64, int64) ([]oathkeeper.Rule, error)
	GetRule(context.Context, string) ([]oathkeeper.Rule, error)
	UpdateRule(context.Context, string, oathkeeper.Rule) error
	CreateRule(context.Context, oathkeeper.Rule) error
	DeleteRule(context.Context, string) error
}
