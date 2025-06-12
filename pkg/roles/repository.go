// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
)

type RoleRepository struct {
	db storage.DBClientInterface

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (r *RoleRepository) FindRoleByName(ctx context.Context, roleName string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.FindRoleByName")
	defer span.End()

	row := r.db.Statement().Select("id", "name", "owner").
		From("role").
		Where(sq.Eq{"name": roleName}).
		QueryRowContext(ctx)

	return findRoleByQuery("FindRoleByName", row)
}

func (r *RoleRepository) FindRoleByIdAndOwner(ctx context.Context, ID, userID string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.FindRoleByIdAndOwner")
	defer span.End()

	row := r.db.Statement().Select("id", "name", "owner").
		From("role").
		Where(sq.Eq{"id": ID}).
		Where(sq.Eq{"owner": userID}).
		QueryRowContext(ctx)

	return findRoleByQuery("FindRoleByIdAndOwner", row)
}

func (r *RoleRepository) FindRoleByNameAndOwner(ctx context.Context, roleName, userID string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.FindRoleByNameAndOwner")
	defer span.End()

	row := r.db.Statement().Select("id", "name", "owner").
		From("role").
		Where(sq.Eq{"name": roleName}).
		Where(sq.Eq{"owner": userID}).
		QueryRowContext(ctx)

	return findRoleByQuery("FindRoleByNameAndOwner", row)
}

func findRoleByQuery(queryName string, row sq.RowScanner) (*Role, error) {
	var id, name, owner string

	err := row.Scan(&id, &name, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("unable to scan %s result, %w", queryName, err)
	}

	return &Role{
		ID:    id,
		Name:  name,
		Owner: owner,
	}, nil
}

func (r *RoleRepository) ListRoles(ctx context.Context, userID string, page, size int64) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.ListRoles")
	defer span.End()

	pageSize := storage.PageSize(size)
	offset := storage.Offset(page, pageSize)

	rows, err := r.db.Statement().Select("name").
		From("role").
		Where(sq.Eq{"owner": userID}).
		Limit(pageSize).
		Offset(offset).
		QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to list roles, %w", err)
	}

	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("unable to scan ListRoles result, %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *RoleRepository) ListRoleGroups(ctx context.Context, roleID string, page, size int64) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.ListRoleGroups")
	defer span.End()

	pageSize := storage.PageSize(size)
	offset := storage.Offset(page, pageSize)

	rows, err := r.db.Statement().Select("name").
		From(`"group" AS g`).
		InnerJoin("group_role AS gr ON gr.group_id = g.id").
		Where(sq.Eq{"role_id": roleID}).
		Limit(pageSize).
		Offset(offset).
		QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to list role groups, %w", err)
	}

	defer rows.Close()

	groups := make([]string, 0)
	for rows.Next() {
		var group string
		if err := rows.Scan(&group); err != nil {
			return nil, fmt.Errorf("unable to scan ListRoleGroups result, %w", err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func (r *RoleRepository) CreateRole(ctx context.Context, roleName, userID string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.CreateRole")
	defer span.End()

	return createRoleQuery(ctx, r.db.Statement(), roleName, userID)
}

func (r *RoleRepository) CreateRoleTx(ctx context.Context, roleName, userID string) (*Role, storage.TxInterface, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.CreateRoleTx")
	defer span.End()

	tx, st, err := r.db.TxStatement(ctx)
	if err != nil {
		return nil, nil, err
	}

	role, err := createRoleQuery(ctx, st, roleName, userID)

	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	return role, tx, nil
}

func createRoleQuery(ctx context.Context, st sq.StatementBuilderType, roleName, userID string) (*Role, error) {
	row := st.Insert("role").
		Columns("name", "owner").
		Values(roleName, userID).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	role := Role{Name: roleName, Owner: userID}

	if err := row.Scan(&role.ID); err != nil {
		return nil, fmt.Errorf("unable to scan CreateRole result, %w", err)
	}

	return &role, nil
}

func (r *RoleRepository) DeleteRoleByName(ctx context.Context, roleName string) (string, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.DeleteRoleByName")
	defer span.End()

	return deleteRoleQuery(ctx, r.db.Statement(), roleName)
}

func (r *RoleRepository) DeleteRoleTx(ctx context.Context, roleName string) (string, storage.TxInterface, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.DeleteRoleTx")
	defer span.End()

	tx, st, err := r.db.TxStatement(ctx)
	if err != nil {
		return "", nil, err
	}

	roleId, err := deleteRoleQuery(ctx, st, roleName)
	if err != nil {
		_ = tx.Rollback()
		return "", nil, err
	}

	return roleId, tx, nil
}

func deleteRoleQuery(ctx context.Context, st sq.StatementBuilderType, roleName string) (string, error) {
	row := st.Delete("role").
		Where(sq.Eq{"name": roleName}).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	var deletedId string
	err := row.Scan(&deletedId)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrNotFound
	}

	if err != nil {
		return "", err
	}

	return deletedId, nil
}

func NewRoleRepository(db storage.DBClientInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *RoleRepository {
	r := new(RoleRepository)

	r.db = db

	r.logger = logger
	r.tracer = tracer
	r.monitor = monitor

	return r
}
