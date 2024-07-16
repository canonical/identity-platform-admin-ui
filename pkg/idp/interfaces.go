// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
)

type AuthorizerInterface interface {
	SetCreateProviderEntitlements(context.Context, string) error
	SetDeleteProviderEntitlements(context.Context, string) error
}

type ServiceInterface interface {
	ListResources(context.Context) ([]*Configuration, error)
	GetResource(context.Context, string) ([]*Configuration, error)
	EditResource(context.Context, string, *Configuration) ([]*Configuration, error)
	CreateResource(context.Context, *Configuration) ([]*Configuration, error)
	DeleteResource(context.Context, string) error
}
