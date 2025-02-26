// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package openfga

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const (
	ASSIGNEE_RELATION = "assignee"
	MEMBER_RELATION   = "member"
	CAN_VIEW_RELATION = "can_view"
)

// TODO @shipperizer this is internal material, worth reusing it across the board
// OpenFGAStore is an overarching store object to deal with OpenFGA entities, meant as a low level
// object to perform cross cutting logic only relevant to the application, therefore doesn't deal with
// user interpolations or returns fancy objects, that is offloaded to the service layer favouring reusability
type OpenFGAStore struct {
	ofga OpenFGAClientInterface

	wpool pool.WorkerPoolInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// ListViewableRoles returns all the roles a specific "assignee"able resource (user, group#member, role#assignee) is linked to (using "can_view" OpenFGA relation)
func (s *OpenFGAStore) ListViewableRoles(ctx context.Context, ID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.ListViewableRoles")
	defer span.End()

	roles, err := s.ofga.ListObjects(ctx, ID, CAN_VIEW_RELATION, "role")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return roles, nil
}

// ListAssignedRoles returns all the roles a specific "assignee"able resource (user, group#member, role#assignee) is linked to (using "assignee" OpenFGA relation)
func (s *OpenFGAStore) ListAssignedRoles(ctx context.Context, assigneeID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.ListAssignedRoles")
	defer span.End()

	roles, err := s.ofga.ListObjects(ctx, assigneeID, ASSIGNEE_RELATION, "role")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return roles, nil
}

// ListAssignedGroups returns all the groups a specific user is memeber of (using "member" OpenFGA relation)
func (s *OpenFGAStore) ListAssignedGroups(ctx context.Context, assigneeID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.ListAssignedGroups")
	defer span.End()

	groups, err := s.ofga.ListObjects(ctx, assigneeID, MEMBER_RELATION, "group")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return groups, nil
}

// AssignRoles assigns roles to an "assignee"able resource (user, group#member)
func (s *OpenFGAStore) AssignRoles(ctx context.Context, assigneeID string, roleIDs ...string) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.AssignRoles")
	defer span.End()

	// preemptive check to verify if all roles to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]Tuple, 0)

	for _, roleID := range roleIDs {
		rs = append(rs, *NewTuple(assigneeID, ASSIGNEE_RELATION, roleID))
	}

	err := s.ofga.WriteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// UnassignRoles drops roles from an "assignee"able resource (user, group#member)
func (s *OpenFGAStore) UnassignRoles(ctx context.Context, assigneeID string, roleIDs ...string) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.UnassignRoles")
	defer span.End()

	// preemptive check to verify if all roles to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]Tuple, 0)

	for _, roleID := range roleIDs {
		rs = append(rs, *NewTuple(assigneeID, ASSIGNEE_RELATION, roleID))
	}

	err := s.ofga.DeleteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// AssignGroups assigns groups to an "assignee"able resource (user, group#member)
func (s *OpenFGAStore) AssignGroups(ctx context.Context, assigneeID string, groupIDs ...string) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.AssignGroups")
	defer span.End()

	// preemptive check to verify if all Groups to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]Tuple, 0)

	for _, groupID := range groupIDs {
		rs = append(rs, *NewTuple(assigneeID, MEMBER_RELATION, groupID))
	}

	err := s.ofga.WriteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// UnassignGroups drops Groups from an "assignee"able resource (user, group#member)
func (s *OpenFGAStore) UnassignGroups(ctx context.Context, assigneeID string, groupIDs ...string) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.UnassignGroups")
	defer span.End()

	// preemptive check to verify if all Groups to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]Tuple, 0)

	for _, groupID := range groupIDs {
		rs = append(rs, *NewTuple(assigneeID, MEMBER_RELATION, groupID))
	}

	err := s.ofga.DeleteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// AssignPermissions assigns permissions to an "assignee"able resource (user, group#member, role#assignee)
func (s *OpenFGAStore) AssignPermissions(ctx context.Context, assigneeID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.AssignPermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *NewTuple(assigneeID, p.Relation, p.Object))
	}

	err := s.ofga.WriteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// UnassignPermissions removes permissions from an "assignee"able resource (user, group#member, role#assignee)
func (s *OpenFGAStore) UnassignPermissions(ctx context.Context, assigneeID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.UnassignPermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *NewTuple(assigneeID, p.Relation, p.Object))
	}

	err := s.ofga.DeleteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// ListPermissions returns all the permissions associated to a specific entity
func (s *OpenFGAStore) ListPermissions(ctx context.Context, ID string, continuationTokens map[string]string) ([]Permission, map[string]string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.ListPermissions")
	defer span.End()

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	// https://go.dev/ref/spec#Send_statements
	// A send on an unbuffered channel can proceed if a receiver is ready.
	// A send on a buffered channel can proceed if there is room in the buffer
	results := make(chan *pool.Result[any], len(s.permissionTypes()))

	wg := sync.WaitGroup{}
	wg.Add(len(s.permissionTypes()))

	for _, t := range s.permissionTypes() {
		s.wpool.Submit(
			s.listPermissionsFunc(ctx, ID, "", t, continuationTokens[t]),
			results,
			&wg,
		)
	}

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)

	permissions := make([]Permission, 0)
	tMap := make(map[string]string)
	errors := make([]error, 0)

	for r := range results {
		v := r.Value.(listPermissionsResult)
		permissions = append(permissions, v.permissions...)
		tMap[v.ofgaType] = v.token

		if v.err != nil {
			errors = append(errors, v.err)
		}
	}

	if len(errors) == 0 {
		return permissions, tMap, nil
	}

	eMsg := ""

	for n, e := range errors {
		s.logger.Errorf(e.Error())
		eMsg = fmt.Sprintf("%s%v - %s\n", eMsg, n, e.Error())
	}

	return permissions, tMap, fmt.Errorf(eMsg)
}

// ListPermissionsWithFilters returns all the permissions associated to a specific entity
func (s *OpenFGAStore) ListPermissionsWithFilters(ctx context.Context, ID string, opts ...ListPermissionsFiltersInterface) ([]Permission, map[string]string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.ListPermissionsWithFilters")
	defer span.End()

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	// https://go.dev/ref/spec#Send_statements
	// A send on an unbuffered channel can proceed if a receiver is ready.
	// A send on a buffered channel can proceed if there is room in the buffer
	results := make(chan *pool.Result[any], len(s.permissionTypes()))

	ff := new(listPermissionsOpts)

	if len(opts) != 0 {
		ff = s.parseFilters(opts...)
	}

	types := s.permissionTypes()
	tokenMap := make(map[string]string)

	if tm := ff.TokenMap; tm != nil {
		tokenMap = tm
	}

	if tf := ff.TypesFilter; len(tf) > 0 {
		types = tf
	}

	wg := sync.WaitGroup{}
	wg.Add(len(types))

	for _, t := range types {
		token, ok := tokenMap[t]

		if !ok {
			token = ""
		}

		s.wpool.Submit(
			s.listPermissionsFunc(ctx, ID, ff.RelationFilter, t, token),
			results,
			&wg,
		)
	}

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)

	permissions := make([]Permission, 0)
	tMap := make(map[string]string)
	errors := make([]error, 0)

	for r := range results {
		v := r.Value.(listPermissionsResult)
		permissions = append(permissions, v.permissions...)
		tMap[v.ofgaType] = v.token

		if v.err != nil {
			errors = append(errors, v.err)
		}
	}

	if len(errors) == 0 {
		return permissions, tMap, nil
	}

	eMsg := ""

	for n, e := range errors {
		s.logger.Errorf(e.Error())
		eMsg = fmt.Sprintf("%s%v - %s\n", eMsg, n, e.Error())
	}

	return permissions, tMap, fmt.Errorf(eMsg)
}

func (s *OpenFGAStore) listPermissionsFunc(ctx context.Context, ID, relation, ofgaType, cToken string) func() any {
	return func() any {
		p, token, err := s.listPermissionsByType(
			ctx,
			ID,
			relation,
			ofgaType,
			cToken,
		)

		return listPermissionsResult{
			permissions: p,
			ofgaType:    ofgaType,
			token:       token,
			err:         err,
		}
	}
}

func (s *OpenFGAStore) listPermissionsByType(ctx context.Context, ID, relation, pType, continuationToken string) ([]Permission, string, error) {
	ctx, span := s.tracer.Start(ctx, "openfga.OpenFGAStore.listPermissionsByType")
	defer span.End()

	r, err := s.ofga.ReadTuples(ctx, ID, relation, fmt.Sprintf("%s:", pType), continuationToken)

	if err != nil {
		s.logger.Error(err.Error())
		return nil, "", err
	}

	permissions := make([]Permission, 0)

	for _, t := range r.GetTuples() {
		// if relation doesn't start with can_ it means it's not a permission (see #assignee)
		if !strings.HasPrefix(t.Key.Relation, "can_") {
			continue
		}

		permissions = append(permissions, Permission{Relation: t.Key.Relation, Object: t.Key.Object})
	}

	return permissions, r.GetContinuationToken(), nil
}

func (s *OpenFGAStore) parseFilters(filters ...ListPermissionsFiltersInterface) *listPermissionsOpts {
	opts := new(listPermissionsOpts)
	opts.TokenMap = make(map[string]string)
	opts.TypesFilter = make([]string, 0)

	// this will keep only the latest filter passed in, if 2 type filters are passed, last one is kept
	for _, filter := range filters {
		switch f := filter.(type) {
		case *TypesFilter:
			if f == nil {
				continue
			}

			if v, ok := f.WithFilter().([]string); ok {
				opts.TypesFilter = v
			} else {
				s.logger.Errorf("wrong types filter, casting failed: %v", f)
			}
		case *RelationFilter:
			if f == nil {
				continue
			}

			if v, ok := f.WithFilter().(string); ok {
				opts.RelationFilter = v
			} else {
				s.logger.Errorf("wrong relation filter, casting failed: %s", f)
			}
		case *TokenMapFilter:
			if f == nil {
				continue
			}

			if v, ok := f.WithFilter().(map[string]string); ok {
				opts.TokenMap = v
			} else {
				s.logger.Errorf("wrong token map, casting failed: %v", f)
			}
		default:
			continue
		}
	}

	return opts
}

func (s *OpenFGAStore) permissionTypes() []string {
	return []string{"group", "role", "identity", "scheme", "provider", "client"}
}

// NewOpenFGAStore returns the implementation of the store
func NewOpenFGAStore(ofga OpenFGAClientInterface, wpool pool.WorkerPoolInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *OpenFGAStore {
	s := new(OpenFGAStore)

	s.ofga = ofga
	s.wpool = wpool

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
