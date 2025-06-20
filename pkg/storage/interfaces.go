// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type DBClientInterface interface {
	Statement() sq.StatementBuilderType
	TxStatement(context.Context) (TxInterface, sq.StatementBuilderType, error)
	Close()
}

type TxInterface interface {
	Commit() error
	Rollback() error
}
