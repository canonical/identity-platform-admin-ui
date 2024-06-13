// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/base64"

	"github.com/google/uuid"
)

// needed for testing purposes
var uuidNewString = uuid.NewString

type OAuth2Helper struct{}

func (o *OAuth2Helper) RandomURLString() string {
	return base64.RawURLEncoding.EncodeToString([]byte(uuidNewString()))
}

func NewOAuth2Helper() *OAuth2Helper {
	return new(OAuth2Helper)
}
