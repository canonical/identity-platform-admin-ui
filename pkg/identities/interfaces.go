// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"

	kClient "github.com/ory/kratos-client-go"
)

type ServiceInterface interface {
	ListIdentities(context.Context, int64, string, string) (*IdentityData, error)
	GetIdentity(context.Context, string) (*IdentityData, error)
	CreateIdentity(context.Context, *kClient.CreateIdentityBody) (*IdentityData, error)
	UpdateIdentity(context.Context, string, *kClient.UpdateIdentityBody) (*IdentityData, error)
	DeleteIdentity(context.Context, string) (*IdentityData, error)
}
