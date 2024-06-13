// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"encoding/base64"
	"testing"
)

func TestOAuth2Helper_RandomURLString(t *testing.T) {
	o := NewOAuth2Helper()
	originalFunc := uuidNewString
	defer func() {
		uuidNewString = originalFunc
	}()

	uuidNewString = func() string {
		return "mock-uuid"
	}

	expected := base64.RawURLEncoding.EncodeToString([]byte("mock-uuid"))

	if got := o.RandomURLString(); got != expected {
		t.Errorf("RandomURLString() = %v, want %v", got, expected)
	}

}
