// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
)

//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_storage.go -source=../storage/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package groups -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

var (
	db      *sql.DB
	mock    sqlmock.Sqlmock
	mockErr error
)

func TestMain(m *testing.M) {
	db, mock, mockErr = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if mockErr != nil {
		panic(mockErr)
	}
	defer db.Close()

	code := m.Run()
	os.Exit(code)
}

func TestGroupRepository_FindGroupByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().Statement().Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.FindGroupByName").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).AddRow("456", "devs", "user2")

		mock.ExpectQuery(query).WithArgs("devs").WillReturnRows(rows)

		group, err := repo.FindGroupByName(ctx, "devs")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if group == nil {
			t.Fatal("expected group, got nil")
		}

		if group.ID != "456" {
			t.Errorf("expected group.ID = '456', got %q", group.ID)
		}

		if group.Name != "devs" {
			t.Errorf("expected group.Name = 'devs', got %q", group.Name)
		}

		if group.Owner != "user2" {
			t.Errorf("expected group.Owner = 'user2', got %q", group.Owner)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("456")

		mock.ExpectQuery(query).WithArgs("devs").WillReturnRows(rows)

		group, err := repo.FindGroupByName(ctx, "devs")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan FindGroupByName result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group on scan error, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1`

		mock.ExpectQuery(query).WithArgs("devs").WillReturnError(sql.ErrNoRows)

		group, err := repo.FindGroupByName(ctx, "devs")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected not found error, got: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_FindGroupByIdAndOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().
		Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).
		AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.FindGroupByIdAndOwner").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE id = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).
			AddRow("group123", "developers", "user1")

		mock.ExpectQuery(query).
			WithArgs("group123", "user1").
			WillReturnRows(rows)

		group, err := repo.FindGroupByIdAndOwner(ctx, "group123", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if group == nil {
			t.Fatal("expected group, got nil")
		}

		if group.ID != "group123" {
			t.Errorf("expected group.ID = 'group123', got %q", group.ID)
		}

		if group.Name != "developers" {
			t.Errorf("expected group.Name = 'developers', got %q", group.Name)
		}

		if group.Owner != "user1" {
			t.Errorf("expected group.Owner = 'user1', got %q", group.Owner)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE id = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow("group123") // manca "name" e "owner"

		mock.ExpectQuery(query).
			WithArgs("group123", "user1").
			WillReturnRows(rows)

		group, err := repo.FindGroupByIdAndOwner(ctx, "group123", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan FindGroupByIdAndOwner result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group on scan error, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE id = $1 AND owner = $2`

		mock.ExpectQuery(query).
			WithArgs("group999", "userXYZ").
			WillReturnError(sql.ErrNoRows)

		group, err := repo.FindGroupByIdAndOwner(ctx, "group999", "userXYZ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group on not found, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_FindGroupByNameAndOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().
		Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).
		AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.FindGroupByNameAndOwner").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).
			AddRow("group123", "engineering", "user1")

		mock.ExpectQuery(query).
			WithArgs("engineering", "user1").
			WillReturnRows(rows)

		group, err := repo.FindGroupByNameAndOwner(ctx, "engineering", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if group == nil {
			t.Fatal("expected group, got nil")
		}

		if group.ID != "group123" {
			t.Errorf("expected group.ID = 'group123', got %q", group.ID)
		}

		if group.Name != "engineering" {
			t.Errorf("expected group.Name = 'engineering', got %q", group.Name)
		}

		if group.Owner != "user1" {
			t.Errorf("expected group.Owner = 'user1', got %q", group.Owner)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow("group123") // mancano name e owner

		mock.ExpectQuery(query).
			WithArgs("engineering", "user1").
			WillReturnRows(rows)

		group, err := repo.FindGroupByNameAndOwner(ctx, "engineering", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if group != nil {
			t.Errorf("expected nil group on scan error, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		query := `SELECT id, name, owner FROM "group" WHERE name = $1 AND owner = $2`

		mock.ExpectQuery(query).
			WithArgs("nonexistent", "userXYZ").
			WillReturnError(sql.ErrNoRows)

		group, err := repo.FindGroupByNameAndOwner(ctx, "nonexistent", "userXYZ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group on not found, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_ListGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.ListGroups").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT name FROM "group" WHERE owner = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("devs").
			AddRow("qa").
			AddRow("ops")

		mock.ExpectQuery(query).WithArgs("user2").WillReturnRows(rows)

		result, err := repo.ListGroups(ctx, "user2", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []string{"devs", "qa", "ops"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d groups, got %d", len(expected), len(result))
		}

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("expected group %d to be %q, got %q", i, expected[i], result[i])
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT name FROM "group" WHERE owner = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).WithArgs("user2").WillReturnRows(rows)

		result, err := repo.ListGroups(ctx, "user2", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan ListGroups result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result on scan error, got: %v", result)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("No rows", func(t *testing.T) {
		query := `SELECT name FROM "group" WHERE owner = $1 LIMIT 100 OFFSET 0`

		mock.ExpectQuery(query).WithArgs("user2").WillReturnRows(sqlmock.NewRows([]string{}))

		result, err := repo.ListGroups(ctx, "user2", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatalf("unexpected nil slice")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %v", result)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Query error", func(t *testing.T) {
		query := `SELECT name FROM "group" WHERE owner = $1 LIMIT 100 OFFSET 0`
		mock.ExpectQuery(query).WithArgs("user2").
			WillReturnError(fmt.Errorf("db failure"))

		result, err := repo.ListGroups(ctx, "user2", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to list groups") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_CreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().
		Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).
		AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.CreateGroup").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `INSERT INTO "group" (name,owner) VALUES ($1,$2) RETURNING id`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("group123")

		mock.ExpectQuery(query).
			WithArgs("engineering", "user1").
			WillReturnRows(rows)

		group, err := repo.CreateGroup(ctx, "engineering", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if group == nil {
			t.Fatal("expected group, got nil")
		}

		if group.ID != "group123" {
			t.Errorf("expected ID 'group123', got %q", group.ID)
		}

		if group.Name != "engineering" {
			t.Errorf("expected Name 'engineering', got %q", group.Name)
		}

		if group.Owner != "user1" {
			t.Errorf("expected Owner 'user1', got %q", group.Owner)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `INSERT INTO "group" (name,owner) VALUES ($1,$2) RETURNING id`

		mock.ExpectQuery(query).
			WithArgs("engineering", "user1").
			WillReturnError(fmt.Errorf("db failure"))

		group, err := repo.CreateGroup(ctx, "engineering", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan CreateGroup result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group, got: %+v", group)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_CreateGroupTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)
	mockTx := NewMockTxInterface(ctrl)

	st := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.CreateGroupTx").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `INSERT INTO "group" (name,owner) VALUES ($1,$2) RETURNING id`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("group123")
		mock.ExpectQuery(query).WithArgs("engineering", "user1").WillReturnRows(rows)

		group, tx, err := repo.CreateGroupTx(ctx, "engineering", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if group == nil {
			t.Fatal("expected group, got nil")
		}

		if group.ID != "group123" {
			t.Errorf("expected ID 'group123', got %q", group.ID)
		}

		if group.Name != "engineering" {
			t.Errorf("expected Name 'engineering', got %q", group.Name)
		}

		if group.Owner != "user1" {
			t.Errorf("expected Owner 'user1', got %q", group.Owner)
		}

		if tx == nil {
			t.Error("expected transaction, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error triggers rollback", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `INSERT INTO "group" (name,owner) VALUES ($1,$2) RETURNING id`
		mock.ExpectQuery(query).
			WithArgs("engineering", "user1").
			WillReturnError(fmt.Errorf("db failure"))

		mockTx.EXPECT().Rollback().Return(nil)

		group, tx, err := repo.CreateGroupTx(ctx, "engineering", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan CreateGroup result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group, got: %+v", group)
		}

		if tx != nil {
			t.Errorf("expected nil tx, got: %+v", tx)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("TxStatement error", func(t *testing.T) {
		mockDBClient.EXPECT().
			TxStatement(ctx).
			Return(nil, sq.StatementBuilderType{}, fmt.Errorf("tx init failed"))

		group, tx, err := repo.CreateGroupTx(ctx, "engineering", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "tx init failed") {
			t.Errorf("unexpected error message: %v", err)
		}

		if group != nil {
			t.Errorf("expected nil group, got: %+v", group)
		}

		if tx != nil {
			t.Errorf("expected nil tx, got: %+v", tx)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_DeleteGroupByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).AnyTimes()

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.DeleteGroupByName").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `DELETE FROM "group" WHERE name = $1 RETURNING id`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("group123")

		mock.ExpectQuery(query).
			WithArgs("engineering").
			WillReturnRows(rows)

		id, err := repo.DeleteGroupByName(ctx, "engineering")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if id != "group123" {
			t.Errorf("expected 'group123', got %v", id)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		query := `DELETE FROM "group" WHERE name = $1 RETURNING id`

		mock.ExpectQuery(query).
			WithArgs("engineering").
			WillReturnError(sql.ErrNoRows)

		id, err := repo.DeleteGroupByName(ctx, "engineering")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}

		if id != "" {
			t.Errorf("expected empty id, got %v", id)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `DELETE FROM "group" WHERE name = $1 RETURNING id`

		mock.ExpectQuery(query).
			WithArgs("engineering").
			WillReturnError(fmt.Errorf("db failure"))

		id, err := repo.DeleteGroupByName(ctx, "engineering")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "db failure" {
			t.Errorf("unexpected error: %v", err)
		}

		if id != "" {
			t.Errorf("expected empty id, got %v", id)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestGroupRepository_DeleteGroupTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)
	mockTx := NewMockTxInterface(ctrl)

	st := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	repo := GroupRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "groups.GroupRepository.DeleteGroupTx").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `DELETE FROM "group" WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{"id"}).AddRow("group123")

		mock.ExpectQuery(query).
			WithArgs("group123").
			WillReturnRows(row)

		id, tx, err := repo.DeleteGroupTx(ctx, "group123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if id != "group123" {
			t.Errorf("expected ID 'group123', got %q", id)
		}

		if tx == nil {
			t.Error("expected non-nil transaction")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Scan error triggers rollback", func(t *testing.T) {
		query := `DELETE FROM "group" WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{}).AddRow()

		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		mock.ExpectQuery(query).
			WithArgs("group999").
			WillReturnRows(row)

		mockTx.EXPECT().Rollback().Return(nil)

		id, tx, err := repo.DeleteGroupTx(ctx, "group999")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != "" {
			t.Errorf("expected empty ID, got %q", id)
		}

		if tx != nil {
			t.Error("expected nil transaction on failure")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("TxStatement error", func(t *testing.T) {
		mockDBClient.EXPECT().
			TxStatement(ctx).
			Return(nil, st, fmt.Errorf("tx init failed"))

		id, tx, err := repo.DeleteGroupTx(ctx, "group456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != "" {
			t.Errorf("expected empty ID, got %q", id)
		}

		if tx != nil {
			t.Error("expected nil transaction")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}
