// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package storage -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package storage -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package storage -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestDBClient_Run_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.Run").Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	client := &DBClient{
		db:       db,
		dbRunner: db,
		logger:   logger,
		tracer:   tracer,
		monitor:  monitor,
	}

	mockAction := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return &ActionResult{}, nil
	}

	_, err = client.Run(context.TODO(), mockAction)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDBClient_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.Run").Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	client := &DBClient{
		db:       db,
		dbRunner: db,
		logger:   logger,
		tracer:   tracer,
		monitor:  monitor,
	}

	expectedErr := errors.New("query failed")

	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return nil, expectedErr
	}

	_, err = client.Run(context.TODO(), action)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDBClient_RunInTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	mockDb.ExpectBegin()
	mockDb.ExpectCommit()

	client := &DBClient{
		db:       db,
		dbRunner: db,
		logger:   logger,
		tracer:   tracer,
		monitor:  monitor,
	}

	action := func(sb sq.StatementBuilderType) (*ActionResult, error) {
		return &ActionResult{}, nil
	}

	rollbackCalled := false
	rollback := func(err error) { rollbackCalled = true }

	_, err = client.RunInTransaction(context.TODO(), action, rollback)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rollbackCalled {
		t.Fatal("rollback should not have been called")
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBClient_RunInTransaction_ActionFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	mockDb.ExpectBegin()
	mockDb.ExpectRollback()

	client := &DBClient{
		db:       db,
		dbRunner: db,
		logger:   logger,
		tracer:   tracer,
		monitor:  monitor,
	}

	expectedErr := errors.New("tx action failed")
	rollbackCalled := false

	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return nil, expectedErr
	}

	rollback := func(err error) {
		if err.Error() == expectedErr.Error() {
			rollbackCalled = true
		}
	}

	_, err = client.RunInTransaction(context.TODO(), action, rollback)
	if err.Error() != expectedErr.Error() {
		t.Fatalf("expected error: %v, got: %v", expectedErr, err)
	}

	if !rollbackCalled {
		t.Fatal("rollback was not called")
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRunInTransaction_BeginTxFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

	mockDb.ExpectBegin().WillReturnError(errors.New("tx begin failed"))

	client := &DBClient{
		db:      db,
		tracer:  tracer,
		logger:  logger,
		monitor: monitor,
	}

	ctx := context.TODO()
	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return nil, nil
	}

	rollback := func(err error) {}

	_, err = client.RunInTransaction(ctx, action, rollback)
	if err == nil {
		t.Fatal("expected error from RunInTransaction when BeginTx fails")
	}

	if err.Error() != "tx begin failed" {
		t.Fatalf("expected error: tx begin failed, got: %v", err)
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRunInTransaction_CommitFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	mockDb.ExpectBegin()
	mockDb.ExpectCommit().WillReturnError(errors.New("commit failed"))

	client := &DBClient{
		db:      db,
		tracer:  tracer,
		logger:  logger,
		monitor: monitor,
	}

	ctx := context.TODO()
	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return &ActionResult{}, nil
	}
	rollback := func(err error) {}

	_, err = client.RunInTransaction(ctx, action, rollback)
	if err == nil {
		t.Fatal("expected error from RunInTransaction when Commit fails")
	}

	if err.Error() != "commit failed" {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRunInTransaction_ActionFailsAndRollbackFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	mockDb.ExpectBegin()
	mockDb.ExpectRollback().WillReturnError(errors.New("rollback failed"))

	client := &DBClient{
		db:      db,
		tracer:  tracer,
		logger:  logger,
		monitor: monitor,
	}

	ctx := context.TODO()
	expectedErr := errors.New("action failed")

	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		return nil, expectedErr
	}

	rollback := func(err error) {}

	_, err = client.RunInTransaction(ctx, action, rollback)
	if err == nil {
		t.Fatal("expected error from RunInTransaction when action and rollback both fail")
	}

	if err.Error() != "rollback failed" {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRunInTransaction_TestRun_NoActionProvided(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.Run").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Error(gomock.Any()).Times(2)

	client := &DBClient{
		db:      db,
		tracer:  tracer,
		logger:  logger,
		monitor: monitor,
	}

	ctx := context.TODO()

	_, err = client.RunInTransaction(ctx, nil, nil)
	if err == nil {
		t.Fatal("expected error from RunInTransaction when action is nil")
	}

	if err.Error() != "query action cannot be null" {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Run(ctx, nil)
	if err == nil {
		t.Fatal("expected error from Run when action is nil")
	}

	if err.Error() != "query action cannot be null" {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestRunInTransaction_PanicRecoveredAndRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	logger := NewMockLoggerInterface(ctrl)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), "storage.DBClient.RunInTransaction").Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	mockDb.ExpectBegin()
	mockDb.ExpectRollback()

	client := &DBClient{
		db:      db,
		tracer:  tracer,
		logger:  logger,
		monitor: monitor,
	}

	ctx := context.TODO()

	action := func(builder sq.StatementBuilderType) (*ActionResult, error) {
		panic("panic in action")
	}

	rollbackCalled := false
	rollback := func(err error) {
		rollbackCalled = true
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatal("panic should have been recovered inside RunInTransaction")
		}
	}()

	_, err = client.RunInTransaction(ctx, action, rollback)
	if err != nil {
		t.Fatalf("RunInTransaction returned unexpected error: %v", err)
	}

	if !rollbackCalled {
		t.Fatal("rollback was not called")
	}

	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
