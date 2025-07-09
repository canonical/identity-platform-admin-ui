// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

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

type GroupRepository struct {
	db storage.DBClientInterface

	logger  logging.LoggerInterface
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
}

func (g *GroupRepository) FindGroupByName(ctx context.Context, groupName string) (*Group, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.FindGroupByName")
	defer span.End()

	row := g.db.Statement().Select("id", "name", "owner").
		From(`"group"`).
		Where(sq.Eq{"name": groupName}).
		QueryRowContext(ctx)

	return findGroupByQuery("FindGroupByName", row)
}

func (g *GroupRepository) FindGroupByIdAndOwner(ctx context.Context, groupID, userID string) (*Group, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.FindGroupByIdAndOwner")
	defer span.End()

	row := g.db.Statement().Select("id", "name", "owner").
		From(`"group"`).
		Where(sq.Eq{"id": groupID}).
		Where(sq.Eq{"owner": userID}).
		QueryRowContext(ctx)

	return findGroupByQuery("FindGroupByIdAndOwner", row)
}

func (g *GroupRepository) FindGroupByNameAndOwner(ctx context.Context, groupName, userID string) (*Group, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.FindGroupByNameAndOwner")
	defer span.End()

	row := g.db.Statement().Select("id", "name", "owner").
		From(`"group"`).
		Where(sq.Eq{"name": groupName}).
		Where(sq.Eq{"owner": userID}).
		QueryRowContext(ctx)

	return findGroupByQuery("FindGroupByNameAndOwner", row)
}

func findGroupByQuery(queryName string, row sq.RowScanner) (*Group, error) {
	var id, name, owner string

	err := row.Scan(&id, &name, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("unable to scan %s result, %w", queryName, err)
	}

	return &Group{
		ID:    id,
		Name:  name,
		Owner: owner,
	}, nil
}

func (g *GroupRepository) ListGroups(ctx context.Context, userID string, page, size int64) ([]string, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.ListGroups")
	defer span.End()

	pageSize := storage.PageSize(size)
	offset := storage.Offset(page, pageSize)

	rows, err := g.db.Statement().Select("name").
		From(`"group"`).
		Where(sq.Eq{"owner": userID}).
		Limit(pageSize).
		Offset(offset).
		QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to list groups, %w", err)
	}

	defer rows.Close()

	groups := make([]string, 0, pageSize)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("unable to scan ListGroups result, %w", err)
		}
		groups = append(groups, name)
	}

	return groups, nil
}

func (g *GroupRepository) CreateGroup(ctx context.Context, groupName, userID string) (*Group, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.CreateGroup")
	defer span.End()

	return createGroupQuery(ctx, g.db.Statement(), groupName, userID)
}

func (g *GroupRepository) CreateGroupTx(ctx context.Context, groupName, userID string) (*Group, storage.TxInterface, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.CreateGroupTx")
	defer span.End()

	tx, st, err := g.db.TxStatement(ctx)
	if err != nil {
		return nil, nil, err
	}

	group, err := createGroupQuery(ctx, st, groupName, userID)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	return group, tx, nil
}

func createGroupQuery(ctx context.Context, st sq.StatementBuilderType, groupName, userID string) (*Group, error) {
	row := st.Insert(`"group"`).
		Columns("name", "owner").
		Values(groupName, userID).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	group := Group{Name: groupName, Owner: userID}

	if err := row.Scan(&group.ID); err != nil {
		return nil, fmt.Errorf("unable to scan CreateGroup result, %w", err)
	}

	return &group, nil
}

func (g *GroupRepository) DeleteGroupByName(ctx context.Context, groupName string) (string, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.DeleteGroupByName")
	defer span.End()

	return deleteGroupQuery(ctx, g.db.Statement(), groupName)
}

func (g *GroupRepository) DeleteGroupTx(ctx context.Context, groupName string) (string, storage.TxInterface, error) {
	ctx, span := g.tracer.Start(ctx, "groups.GroupRepository.DeleteGroupTx")
	defer span.End()

	tx, st, err := g.db.TxStatement(ctx)
	if err != nil {
		return "", nil, err
	}

	groupID, err := deleteGroupQuery(ctx, st, groupName)
	if err != nil {
		_ = tx.Rollback()
		return "", nil, err
	}

	return groupID, tx, nil
}

func deleteGroupQuery(ctx context.Context, st sq.StatementBuilderType, groupName string) (string, error) {
	row := st.Delete(`"group"`).
		Where(sq.Eq{"name": groupName}).
		Suffix("RETURNING id").
		QueryRowContext(ctx)

	var deletedID string
	err := row.Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return deletedID, nil
}

func NewGroupRepository(db storage.DBClientInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *GroupRepository {
	r := new(GroupRepository)

	r.db = db

	r.logger = logger
	r.tracer = tracer
	r.monitor = monitor

	return r
}
