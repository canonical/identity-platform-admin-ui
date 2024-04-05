// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package schemas

import (
	"context"

	kClient "github.com/ory/kratos-client-go"
)

type ServiceInterface interface {
	ListSchemas(context.Context, int64, string) (*IdentitySchemaData, error)
	GetSchema(context.Context, string) (*IdentitySchemaData, error)
	EditSchema(context.Context, string, *kClient.IdentitySchemaContainer) (*IdentitySchemaData, error)
	CreateSchema(context.Context, *kClient.IdentitySchemaContainer) (*IdentitySchemaData, error)
	DeleteSchema(context.Context, string) error
	GetDefaultSchema(context.Context) (*DefaultSchema, error)
	UpdateDefaultSchema(context.Context, *DefaultSchema) (*DefaultSchema, error)
}
