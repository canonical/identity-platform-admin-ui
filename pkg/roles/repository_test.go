// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
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

func TestRoleRepository_FindRole(t *testing.T) {
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
		Return(sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)).Times(3)

	repo := RoleRepository{
		db:      mockDBClient,
		logger:  mockLogger,
		tracer:  mockTracer,
		monitor: mockMonitor,
	}

	ctx := context.TODO()

	mockTracer.EXPECT().
		Start(gomock.Any(), "roles.RoleRepository.GetRole").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {

		query := `SELECT id, name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("123", "admin")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		role, err := repo.FindRole(ctx, "user1", "123")
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
	})

	t.Run("Scan error", func(t *testing.T) {

		query := `SELECT id, name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{"id"}).AddRow("123")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		role, err := repo.FindRole(ctx, "user1", "123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "unable to scan FindRole result") {
			t.Errorf("unexpected error message: %v", err)
		}

		if role != nil {
			t.Errorf("expected nil role on scan error, got: %+v", role)
		}
	})

	t.Run("Not found", func(t *testing.T) {

		query := `SELECT id, name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("", "")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		role, err := repo.FindRole(ctx, "user1", "123")
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
		query := `SELECT name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("admin").
			AddRow("editor").
			AddRow("viewer")

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		result, err := repo.ListRoles(ctx, "user1")
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
		query := `SELECT name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		result, err := repo.ListRoles(ctx, "user1")
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
		query := `SELECT name FROM role WHERE owner = $1`
		rows := sqlmock.NewRows([]string{"name"})

		mock.ExpectQuery(query).WithArgs("user1").WillReturnRows(rows)

		result, err := repo.ListRoles(ctx, "user1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected empty slice, got nil")
		}

		if len(result) != 0 {
			t.Errorf("expected 0 roles, got %d", len(result))
		}
	})

	t.Run("Query error", func(t *testing.T) {
		query := `SELECT name FROM role WHERE owner = $1`
		mock.ExpectQuery(query).WithArgs("user1").
			WillReturnError(fmt.Errorf("db failure"))

		result, err := repo.ListRoles(ctx, "user1")
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
		query := `SELECT group.name FROM group INNER JOIN group_role WHERE group_role.role_id = $1 AND group_role.group_id = $2`
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("admins").
			AddRow("editors")

		mock.ExpectQuery(query).WithArgs("role123", "group.id").WillReturnRows(rows)

		result, err := repo.ListRoleGroups(ctx, "role123")
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
		query := `SELECT group.name FROM group INNER JOIN group_role WHERE group_role.role_id = $1 AND group_role.group_id = $2`
		rows := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).WithArgs("role123", "group.id").WillReturnRows(rows)

		result, err := repo.ListRoleGroups(ctx, "role123")
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
		query := `SELECT group.name FROM group INNER JOIN group_role WHERE group_role.role_id = $1 AND group_role.group_id = $2`
		rows := sqlmock.NewRows([]string{})

		mock.ExpectQuery(query).WithArgs("role123", "group.id").WillReturnRows(rows)

		result, err := repo.ListRoleGroups(ctx, "role123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected empty slice, got nil")
		}

		if len(result) != 0 {
			t.Errorf("expected 0 groups, got %d", len(result))
		}
	})

	t.Run("Query error", func(t *testing.T) {
		query := `SELECT group.name FROM group INNER JOIN group_role WHERE group_role.role_id = $1 AND group_role.group_id = $2`
		mock.ExpectQuery(query).WithArgs("role123", "group.id").
			WillReturnError(fmt.Errorf("db failure"))

		result, err := repo.ListRoleGroups(ctx, "role123")
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

		role, err := repo.CreateRole(ctx, "user1", "admin")
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

		role, err := repo.CreateRole(ctx, "user1", "admin")
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
		Start(gomock.Any(), "roles.RoleRepository.DeleteRole").
		Return(ctx, trace.SpanFromContext(context.TODO())).
		AnyTimes()

	t.Run("Success", func(t *testing.T) {
		query := `DELETE FROM role WHERE owner = $1 AND id = $2 RETURNING id`
		row := sqlmock.NewRows([]string{"id"}).AddRow(123)

		mock.ExpectQuery(query).
			WithArgs("user123", "role123").
			WillReturnRows(row)

		id, err := repo.DeleteRole(ctx, "user123", "role123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if id != 123 {
			t.Errorf("expected 123, got %v", id)
		}
	})

	t.Run("Scan error", func(t *testing.T) {
		query := `DELETE FROM role WHERE owner = $1 AND id = $2 RETURNING id`
		row := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).
			WithArgs("user123", "role123").
			WillReturnRows(row)

		id, err := repo.DeleteRole(ctx, "user123", "role123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != 0 {
			t.Errorf("expected empty id, got %v", id)
		}
	})

	/*t.Run("Not found", func(t *testing.T) {
		query := `DELETE FROM role WHERE owner = $1 AND id = $2 RETURNING id`
		row := sqlmock.NewRows([]string{}).AddRow()

		mock.ExpectQuery(query).
			WithArgs("user123", "role123").
			WillReturnRows(row)

		id, err := repo.DeleteRole(ctx, "user123", "role123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if id != 0 {
			t.Errorf("expected empty id, got %v", id)
		}
	})*/
}
