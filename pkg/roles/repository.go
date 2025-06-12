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

func (r *RoleRepository) FindRole(ctx context.Context, userID, ID string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.GetRole")
	defer span.End()

	row := r.db.Statement().Select("id", "name").
		From(storage.ROLE_TABLE).
		Where(sq.Eq{storage.ROLE_ID: ID}).
		Where(sq.Eq{storage.ROLE_OWNER: userID}).
		QueryRowContext(ctx)

	var id, name string

	err := row.Scan(&id, &name)
	if err != nil {
		return nil, fmt.Errorf("unable to scan FindRole result, %w", err)
	}

	if id == "" {
		return nil, storage.ErrNotFound
	}

	return &Role{
		ID:   id,
		Name: name,
	}, nil
}

func (r *RoleRepository) ListRoles(ctx context.Context, userID string) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.ListRoles")
	defer span.End()

	rows, err := r.db.Statement().Select("name").
		From(storage.ROLE_TABLE).
		Where(sq.Eq{storage.ROLE_OWNER: userID}).
		QueryContext(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return []string{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to list roles, %w", err)
	}

	defer func() { _ = rows.Close() }()

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

func (r *RoleRepository) ListRoleGroups(ctx context.Context, roleID string) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.ListRoleGroups")
	defer span.End()

	rows, err := r.db.Statement().Select(storage.GROUP_NAME).
		From(storage.GROUP_TABLE).
		InnerJoin(storage.GROUP_ROLE_TABLE).
		Where(sq.Eq{storage.GROUP_ROLE_ROLE_ID: roleID}).
		Where(sq.Eq{storage.GROUP_ROLE_GROUP_ID: storage.GROUP_ID}).
		QueryContext(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return []string{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to list role groups, %w", err)
	}

	defer func() { _ = rows.Close() }()

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

func (r *RoleRepository) CreateRole(ctx context.Context, userID, roleName string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.CreateRole")
	defer span.End()

	row := r.db.Statement().Insert(storage.ROLE_TABLE).
		Columns(storage.ROLE_NAME, storage.ROLE_OWNER).
		Values(roleName, userID).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	role := Role{Name: roleName}

	if err := row.Scan(&role.ID); err != nil {
		return nil, fmt.Errorf("unable to scan CreateRole result, %w", err)
	}

	return &role, nil
}

func (r *RoleRepository) DeleteRole(ctx context.Context, userID, roleID string) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "roles.RoleRepository.DeleteRole")
	defer span.End()

	row := r.db.Statement().Delete(storage.ROLE_TABLE).
		Where(sq.Eq{storage.ROLE_ID: roleID}).
		Where(sq.Eq{storage.ROLE_OWNER: userID}).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	var deletedId int64 = -1
	err := row.Scan(&deletedId)
	if err != nil {
		return 0, err
	}

	if deletedId == -1 {
		return 0, storage.ErrNotFound
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
