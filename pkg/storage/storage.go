// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const (
	defaultPage     uint64 = 1
	defaultPageSize uint64 = 100
)

func Offset(pageParam int64, pageSize uint64) uint64 {
	if pageParam <= 0 {
		return (defaultPage - 1) * pageSize
	}
	return uint64(pageParam) * pageSize
}

func PageSize(sizeParam int64) uint64 {
	if sizeParam <= 0 {
		return defaultPageSize
	}
	return uint64(sizeParam)
}

var ErrNotFound = fmt.Errorf("storage: resource not found")

type DBClient struct {
	// pool is the native PGX pool we hold to allow closing
	pool *pgxpool.Pool
	// db original instance to handle transactions
	db *sql.DB
	// dbRunner is the runner instance of choice (either original DB or db with query cache, cannot be used for transactions)
	dbRunner sq.BaseRunner

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (d *DBClient) Statement() sq.StatementBuilderType {
	return sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		RunWith(d.dbRunner)
}

func (d *DBClient) TxStatement(ctx context.Context) (TxInterface, sq.StatementBuilderType, error) {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		return nil, sq.StatementBuilderType{}, err
	}

	return tx, sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx), nil
}

func (d *DBClient) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}

	if d.pool != nil {
		d.pool.Close()
	}
}

func NewDBClient(dsn string, queryCacheEnabled bool, tracingEnabled bool, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *DBClient {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Fatalf("DSN validation failed, shutting down, err: %v", err)
	}

	if tracingEnabled {
		// otelpgx.NewTracer will use default global TracerProvider, just like our tracer struct
		config.ConnConfig.Tracer = otelpgx.NewTracer()
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Fatalf("DB pool creation failed, shutting down, err: %v", err)
	}

	if tracingEnabled {
		// when tracing is enabled, also collect metrics
		if err := otelpgx.RecordStats(pool); err != nil {
			logger.Fatalf("unable to start metrics collection for database: %v", err)
		}
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := db.Ping(); err != nil {
		logger.Fatalf("DB connection failed, shutting down, err: %v", err)
	}

	d := new(DBClient)
	d.pool = pool
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
