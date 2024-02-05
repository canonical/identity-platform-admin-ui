// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package logging

import (
	"go.uber.org/zap"
)

func NewNoopLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}
