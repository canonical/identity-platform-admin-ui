// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type DBClientInterface interface {
	Statement() sq.StatementBuilderType
	RunInTransaction(ctx context.Context, action QueryAction, rollback RollbackAction) (*ActionResult, error)
	Close()
}
