// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type QueryAction func(builderType sq.StatementBuilderType) (*ActionResult, error)
type RollbackAction func(error)

var (
	NoopAction   QueryAction    = func(builderType sq.StatementBuilderType) (*ActionResult, error) { return nil, nil }
	NoopRollback RollbackAction = func(error) {}
)

type ActionResult struct {
	Rows *sql.Rows
	Row  sq.RowScanner
}

func (a *ActionResult) hasSingleRow() bool {
	return a.Row != nil
}

func (a *ActionResult) hasMultipleRows() bool {
	return a.Rows != nil
}

func NewMultipleActionResult(rows *sql.Rows) *ActionResult {
	return &ActionResult{Rows: rows}
}

func NewSingleActionResult(row sq.RowScanner) *ActionResult {
	return &ActionResult{Row: row}
}

type DBClient struct {
	// db original instance to handle transactions
	db *sql.DB
	// dbRunner is the runner instance of choice (either original DB or db with query cache, cannot be used for transactions)
	dbRunner sq.BaseRunner

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (d *DBClient) Run(ctx context.Context, action QueryAction) (*ActionResult, error) {
	ctx, span := d.tracer.Start(ctx, "storage.DBClient.Run")
	defer span.End()

	if action == nil {
		d.logger.Error("query action cannot be null")
		return nil, fmt.Errorf("query action cannot be null")
	}

	statementBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(d.dbRunner)

	ret, err := action(statementBuilder)
	if err != nil {
		d.logger.Errorf("Failed to execute QueryAction, err: %v", err)
		return nil, err
	}

	return ret, nil
}

func (d *DBClient) RunInTransaction(ctx context.Context, action QueryAction, rollback RollbackAction) (*ActionResult, error) {
	ctx, span := d.tracer.Start(ctx, "storage.DBClient.RunInTransaction")
	defer span.End()

	if action == nil {
		d.logger.Error("query action cannot be null")
		return nil, fmt.Errorf("query action cannot be null")
	}

	if rollback == nil {
		rollback = NoopRollback
	}

	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		d.logger.Errorf("Failed to begin transaction, err: %v", err)
		return nil, err
	}

	statementBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx)

	defer func() {
		if err := recover(); err != nil {
			d.logger.Errorf("Recovered from panic (rolling back tx): %v", err)
			rollback(fmt.Errorf("recovered from panic: %v", err))
			_ = tx.Rollback()
		}
	}()

	ret, actionErr := action(statementBuilder)
	if actionErr != nil {
		d.logger.Errorf("Failed to execute QueryAction, err: %v", actionErr)

		rollback(actionErr)

		if err := tx.Rollback(); err != nil {
			d.logger.Errorf("Failed to rollback transaction, err: %v", err)
			return nil, err
		}

		return nil, actionErr
	}

	if err := tx.Commit(); err != nil {
		d.logger.Errorf("Failed to commit transaction, err: %v", err)
		return nil, err
	}

	return ret, nil
}

func (d *DBClient) Close() error {
	if d.db != nil {
		return d.db.Close()
	}

	return nil
}

func NewDBClient(dsn string, queryCacheEnabled bool, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *DBClient {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		logger.Fatalf("DSN validation failed, shutting down, err: %v", err)
	}

	db := stdlib.OpenDB(*config)
	if err := db.Ping(); err != nil {
		logger.Fatalf("DB connection failed, shutting down, err: %v", err)
	}

	d := new(DBClient)
	d.db = db
	d.dbRunner = db

	if queryCacheEnabled {
		d.dbRunner = sq.NewStmtCache(db)
	}

	d.tracer = tracer
	d.monitor = monitor
	d.logger = logger

	return d
}
