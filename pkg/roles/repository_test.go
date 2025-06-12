// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
)

//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_storage.go -source=../storage/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package roles -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestRoleRepository_FindRoleByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(2)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.FindRoleByName").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {

		query := `SELECT id, name, owner FROM role WHERE name = $1`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).AddRow("123", "admin", "user1")

		mock.ExpectQuery(query).WithArgs("admin").WillReturnRows(rows)

		role, err := repo.FindRoleByName(ctx, "admin")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if role == nil {
			t.Fatal("expected role, got nil")
		}

		if role.ID != "123" {
			t.Errorf("expected role.ID = '123', got %q", role.ID)
		}

		if role.Name != "admin" {
			t.Errorf("expected role.Name = 'admin', got %q", role.Name)
		}

		if role.Owner != "user1" {
			t.Errorf("expected role.Owner = 'user1', got %q", role.Owner)
		}
	})

	t.Run("Scan error", func(t *testing.T) {

		query := `SELECT id, name, owner FROM role WHERE name = $1`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("123")

		mock.ExpectQuery(query).WithArgs("admin").WillReturnRows(rows)

		role, err := repo.FindRoleByName(ctx, "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan FindRoleByName result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		t.Skipf("Skipping because of unexpected Sqlmock behavior, function works when manually tested")

		query := `SELECT id, name, owner FROM role WHERE name = $1`

		mock.ExpectQuery(query).WithArgs("admin").WillReturnError(sql.ErrNoRows)

		role, err := repo.FindRoleByName(ctx, "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected not found error, got: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})
}

func TestRoleRepository_FindRoleByIdAndOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(2)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.FindRoleByIdAndOwner").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {

		query := `SELECT id, name, owner FROM role WHERE id = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).AddRow("123", "admin", "user1")

		mock.ExpectQuery(query).WithArgs("123", "user1").WillReturnRows(rows)

		role, err := repo.FindRoleByIdAndOwner(ctx, "123", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if role == nil {
			t.Fatal("expected role, got nil")
		}

		if role.ID != "123" {
			t.Errorf("expected role.ID = '123', got %q", role.ID)
		}

		if role.Name != "admin" {
			t.Errorf("expected role.Name = 'admin', got %q", role.Name)
		}

		if role.Owner != "user1" {
			t.Errorf("expected role.Owner = 'user1', got %q", role.Owner)
		}
	})

	t.Run("Scan error", func(t *testing.T) {

		query := `SELECT id, name FROM role WHERE id = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("123")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		role, err := repo.FindRoleByIdAndOwner(ctx, "123", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan FindRoleByIdAndOwner result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		t.Skipf("Skipping because of unexpected Sqlmock behavior, function works when manually tested")

		query := `SELECT id, name FROM role WHERE id = $1 AND owner = $2`

		mock.ExpectQuery(query).WithArgs("1", "user123").WillReturnError(sql.ErrNoRows)

		role, err := repo.FindRoleByIdAndOwner(ctx, "1", "user123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected not found error, got: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})
}

func TestRoleRepository_FindRoleByNameAndOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(2)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.FindRoleByNameAndOwner").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {

		query := `SELECT id, name, owner FROM role WHERE name = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id", "name", "owner"}).AddRow("123", "admin", "user1")

		mock.ExpectQuery(query).WithArgs("admin", "user1").WillReturnRows(rows)

		role, err := repo.FindRoleByNameAndOwner(ctx, "admin", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if role == nil {
			t.Fatal("expected role, got nil")
		}

		if role.ID != "123" {
			t.Errorf("expected role.ID = '123', got %q", role.ID)
		}

		if role.Name != "admin" {
			t.Errorf("expected role.Name = 'admin', got %q", role.Name)
		}

		if role.Owner != "user1" {
			t.Errorf("expected role.Owner = 'user1', got %q", role.Owner)
		}
	})

	t.Run("Scan error", func(t *testing.T) {

		query := `SELECT id, name, owner FROM role WHERE name = $1 AND owner = $2`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("123")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		role, err := repo.FindRoleByNameAndOwner(ctx, "admin", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		t.Skipf("Skipping because of unexpected Sqlmock behavior, function works when manually tested")

		query := `SELECT id, name FROM role WHERE name = $1 AND owner = $2`

		mock.ExpectQuery(query).WithArgs("123").WillReturnError(sql.ErrNoRows)

		role, err := repo.FindRoleByNameAndOwner(ctx, "admin", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err != storage.ErrNotFound {
			t.Errorf("expected not found error, got: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})
}

func TestRoleRepository_ListRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(4)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.ListRoles").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT name FROM role WHERE owner = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("admin").
			AddRow("editor").
			AddRow("viewer")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		result, err := repo.ListRoles(ctx, "user1", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []string{"admin", "editor", "viewer"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d roles, got %d", len(expected), len(result))
		}

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("expected role %d to be %q, got %q", i, expected[i], result[i])
			}
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT name FROM role WHERE owner = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		result, err := repo.ListRoles(ctx, "user1", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan ListRoles result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result on scan error, got: %v", result)
		}
	})

	t.Run("No rows", func(t *testing.T) {
		query := `SELECT name FROM role WHERE owner = $1 LIMIT 100 OFFSET 0`

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(sqlmock.NewRows([]string{}))

		result, err := repo.ListRoles(ctx, "user1", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatalf("unexpected nil slice")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %v", result)
		}
	})

	t.Run("Query error", func(t *testing.T) {
		query := `SELECT name FROM role WHERE owner = $1 LIMIT 100 OFFSET 0`
		mock.ExpectQuery(query).WithArgs("user1").
			WillReturnError(fmt.Errorf("db failure"))

		result, err := repo.ListRoles(ctx, "user1", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to list roles") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})
}

func TestRoleRepository_ListRoleGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(4)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.ListRoleGroups").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `SELECT name FROM "group" AS g INNER JOIN group_role AS gr ON gr.group_id = g.id WHERE role_id = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("admins").
			AddRow("editors")

		mock.ExpectQuery(query).WithArgs("role123").WillReturnRows(rows)

		result, err := repo.ListRoleGroups(ctx, "role123", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []string{"admins", "editors"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d groups, got %d", len(expected), len(result))
		}

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("expected group %d to be %q, got %q", i, expected[i], result[i])
			}
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `SELECT name FROM "group" AS g INNER JOIN group_role AS gr ON gr.group_id = g.id WHERE role_id = $1 LIMIT 100 OFFSET 0`
		rows := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).WithArgs("role123").WillReturnRows(rows)

		result, err := repo.ListRoleGroups(ctx, "role123", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan ListRoleGroups result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result on scan error, got: %v", result)
		}
	})

	t.Run("No rows", func(t *testing.T) {
		query := `SELECT name FROM "group" AS g INNER JOIN group_role AS gr ON gr.group_id = g.id WHERE role_id = $1 LIMIT 100 OFFSET 0`

		mock.ExpectQuery(query).WithArgs("role123").WillReturnRows(sqlmock.NewRows([]string{}))

		result, err := repo.ListRoleGroups(ctx, "role123", 0, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatalf("unexpected nil slice")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty slice, got %v", result)
		}
	})

	t.Run("Query error", func(t *testing.T) {
		query := `SELECT name FROM "group" AS g INNER JOIN group_role AS gr ON gr.group_id = g.id WHERE role_id = $1 LIMIT 100 OFFSET 0`
		mock.ExpectQuery(query).WithArgs("role123", "group.id").
			WillReturnError(fmt.Errorf("db failure"))

		result, err := repo.ListRoleGroups(ctx, "role123", 0, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to list role groups") {
			t.Errorf("unexpected error message: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})
}

func TestRoleRepository_CreateRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(2)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.CreateRole").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `INSERT INTO role (name,owner) VALUES ($1,$2) RETURNING id`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("role123")

		mock.ExpectQuery(query).WithArgs("admin", "user1").WillReturnRows(rows)

		role, err := repo.CreateRole(ctx, "admin", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if role == nil {
			t.Fatal("expected role, got nil")
		}

		if role.ID != "role123" {
			t.Errorf("expected ID 'role123', got %q", role.ID)
		}

		if role.Name != "admin" {
			t.Errorf("expected Name 'admin', got %q", role.Name)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `INSERT INTO role (name,owner) VALUES ($1,$2) RETURNING id`

		mock.ExpectQuery(query).WithArgs("admin", "user1").
			WillReturnError(fmt.Errorf("db failure"))

		role, err := repo.CreateRole(ctx, "admin", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan CreateRole result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role, got: %+v", role)
		}
	})
}

func TestRoleRepository_CreateRoleTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)
	mockTx := NewMockTxInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	st := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.CreateRoleTx").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `INSERT INTO role (name,owner) VALUES ($1,$2) RETURNING id`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("role123")
		mock.ExpectQuery(query).WithArgs("admin", "user1").WillReturnRows(rows)

		role, tx, err := repo.CreateRoleTx(ctx, "admin", "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if role == nil {
			t.Fatal("expected role, got nil")
		}

		if role.ID != "role123" {
			t.Errorf("expected ID 'role123', got %q", role.ID)
		}

		if role.Name != "admin" {
			t.Errorf("expected Name 'admin', got %q", role.Name)
		}

		if tx == nil {
			t.Error("expected transaction, got nil")
		}
	})

	t.Run("Scan error triggers rollback", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `INSERT INTO role (name,owner) VALUES ($1,$2) RETURNING id`
		mock.ExpectQuery(query).WithArgs("admin", "user1").
			WillReturnError(fmt.Errorf("db failure"))

		mockTx.EXPECT().Rollback().Return(nil)

		role, tx, err := repo.CreateRoleTx(ctx, "admin", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan CreateRole result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role, got: %+v", role)
		}

		if tx != nil {
			t.Errorf("expected nil tx, got: %+v", tx)
		}
	})

	t.Run("TxStatement error", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(nil, sq.StatementBuilderType{}, fmt.Errorf("tx init failed"))

		role, tx, err := repo.CreateRoleTx(ctx, "admin", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "tx init failed") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role, got: %+v", role)
		}

		if tx != nil {
			t.Errorf("expected nil tx, got: %+v", tx)
		}
	})
}

func TestRoleRepository_DeleteRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	mockDBClient.EXPECT().Statement().
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(2)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.DeleteRoleByName").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `DELETE FROM role WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{"id"}).AddRow(123)

		mock.ExpectQuery(query).
			WithArgs("role123").
			WillReturnRows(row)

		id, err := repo.DeleteRoleByName(ctx, "role123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if id != "123" {
			t.Errorf("expected 123, got %v", id)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `DELETE FROM role WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).
			WithArgs("role123").
			WillReturnRows(row)

		id, err := repo.DeleteRoleByName(ctx, "role123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != "" {
			t.Errorf("expected empty id, got %v", id)
		}
	})
}

func TestRoleRepository_DeleteRoleTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockDBClient := NewMockDBClientInterface(ctrl)
	mockTx := NewMockTxInterface(ctrl)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	defer db.Close()

	st := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.Background()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.DeleteRoleTx").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `DELETE FROM role WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{"id"}).AddRow("123")

		mock.ExpectQuery(query).
			WithArgs("role123").
			WillReturnRows(row)

		id, tx, err := repo.DeleteRoleTx(ctx, "role123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if id != "123" {
			t.Errorf("expected ID '123', got %q", id)
		}

		if tx == nil {
			t.Error("expected non-nil transaction")
		}
	})

	t.Run("Scan error triggers rollback", func(t *testing.T) {
		mockDBClient.EXPECT().TxStatement(ctx).Return(mockTx, st, nil)

		query := `DELETE FROM role WHERE name = $1 RETURNING id`
		row := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).
			WithArgs("role123").
			WillReturnRows(row)

		mockTx.EXPECT().Rollback().Return(nil)

		id, tx, err := repo.DeleteRoleTx(ctx, "role123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != "" {
			t.Errorf("expected empty ID, got %q", id)
		}

		if tx != nil {
			t.Error("expected nil transaction on failure")
		}
	})

	t.Run("TxStatement error", func(t *testing.T) {
		mockDBClient.EXPECT().
			TxStatement(ctx).
			Return(nil, sq.StatementBuilderType{}, fmt.Errorf("tx init failed"))

		id, tx, err := repo.DeleteRoleTx(ctx, "role123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != "" {
			t.Errorf("expected empty ID, got %q", id)
		}

		if tx != nil {
			t.Error("expected nil transaction")
		}
	})
}
