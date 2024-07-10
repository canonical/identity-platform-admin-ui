// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"context"

	kClient "github.com/ory/kratos-client-go"
)

type AuthorizerInterface interface {
	SetCreateSchemaEntitlements(context.Context, string) error
	SetDeleteSchemaEntitlements(context.Context, string) error
}

type ServiceInterface interface {
	ListSchemas(context.Context, int64, string) (*IdentitySchemaData, error)
	GetSchema(context.Context, string) (*IdentitySchemaData, error)
	EditSchema(context.Context, string, *kClient.IdentitySchemaContainer) (*IdentitySchemaData, error)
	CreateSchema(context.Context, *kClient.IdentitySchemaContainer) (*IdentitySchemaData, error)
	DeleteSchema(context.Context, string) error
	GetDefaultSchema(context.Context) (*DefaultSchema, error)
	UpdateDefaultSchema(context.Context, *DefaultSchema) (*DefaultSchema, error)
}
