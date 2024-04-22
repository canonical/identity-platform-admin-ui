// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
)

const (
	MEMBER_RELATION   = "member"
	ASSIGNEE_RELATION = "assignee"
)

type listPermissionsResult struct {
	permissions []string
	token       string
	ofgaType    string
	err         error
}

// Service contains the business logic to deal with groups on the Admin UI OpenFGA model
type Service struct {
	ofga OpenFGAClientInterface

	wpool pool.WorkerPoolInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) buildGroupMember(ctx context.Context, ID string) string {
	_, span := s.tracer.Start(ctx, "groups.Service.buildGroupMember")
	defer span.End()

	return fmt.Sprintf("group:%s#%s", ID, MEMBER_RELATION)

}

// ListGroups returns all the groups a specific user can see (using "can_view" OpenFGA relation)
func (s *Service) ListGroups(ctx context.Context, userID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.ListGroups")
	defer span.End()

	groups, err := s.ofga.ListObjects(ctx, fmt.Sprintf("user:%s", userID), "can_view", "group")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return groups, nil
}

// ListRoles returns all the roles associated to a specific group
func (s *Service) ListRoles(ctx context.Context, ID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.ListRoles")
	defer span.End()

	roles, err := s.ofga.ListObjects(ctx, s.buildGroupMember(ctx, ID), ASSIGNEE_RELATION, "role")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return roles, nil
}

// ListPermissions returns all the permissions associated to a specific group
func (s *Service) ListPermissions(ctx context.Context, ID string, continuationTokens map[string]string) ([]string, map[string]string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.ListPermissions")
	defer span.End()

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	// https://go.dev/ref/spec#Send_statements
	// A send on an unbuffered channel can proceed if a receiver is ready.
	// A send on a buffered channel can proceed if there is room in the buffer
	results := make(chan *pool.Result[any], len(s.permissionTypes()))

	wg := sync.WaitGroup{}
	wg.Add(len(s.permissionTypes()))

	// TODO @shipperizer use a background operator
	for _, t := range s.permissionTypes() {
		s.wpool.Submit(
			s.listPermissionsFunc(ctx, ID, t, continuationTokens[t]),
			results,
			&wg,
		)
	}

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)

	permissions := make([]string, 0)
	tMap := make(map[string]string)
	errors := make([]error, 0)

	for r := range results {
		s.logger.Info(results)
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

// GetGroup returns the specified group using the ID argument, userID is used to validate the visibility by the user
// making the call
func (s *Service) GetGroup(ctx context.Context, userID, ID string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.GetGroup")
	defer span.End()

	exists, err := s.ofga.Check(ctx, fmt.Sprintf("user:%s", userID), "can_view", fmt.Sprintf("group:%s", ID))

	if err != nil {

		s.logger.Error(err.Error())
		return "", err
	}

	if exists {
		return ID, nil
	}

	// if we got here it means authorization check hasn't worked
	return "", nil
}

// CreateGroup creates a group and associates it with the userID passed as argument
// an extra tuple is created to estabilish the "privileged" relatin for admin users
func (s *Service) CreateGroup(ctx context.Context, userID, ID string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.CreateGroup")
	defer span.End()

	// TODO @shipperizer will we need also the can_view?
	// does creating a group mean that you are the owner, therefore u get all the permissions on it?
	// right now assumption is only admins will be able to do this
	// potentially changing the model to say
	// `define can_view: [user, user:*, group#assignee, group#member] or assignee or admin from privileged`
	// might sort the problem

	// TODO @shipperizer offload to privileged creator object
	err := s.ofga.WriteTuples(
		ctx,
		*ofga.NewTuple(fmt.Sprintf("user:%s", userID), MEMBER_RELATION, fmt.Sprintf("group:%s", ID)),
		*ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("group:%s", ID)),
	)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// AssignRoles assigns roles to a group
func (s *Service) AssignRoles(ctx context.Context, ID string, roles ...string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.AssignRoles")
	defer span.End()

	// preemptive check to verify if all roles to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]ofga.Tuple, 0)

	for _, role := range roles {
		rs = append(rs, *ofga.NewTuple(s.buildGroupMember(ctx, ID), ASSIGNEE_RELATION, fmt.Sprintf("role:%s", role)))
	}

	err := s.ofga.WriteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// RemoveRoles drops roles from a group
func (s *Service) RemoveRoles(ctx context.Context, ID string, roles ...string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.RemoveRoles")
	defer span.End()

	// preemptive check to verify if all roles to be assigned are accessible by the user
	// needs to happen separately

	rs := make([]ofga.Tuple, 0)

	for _, role := range roles {
		rs = append(rs, *ofga.NewTuple(s.buildGroupMember(ctx, ID), ASSIGNEE_RELATION, fmt.Sprintf("role:%s", role)))
	}

	err := s.ofga.DeleteTuples(ctx, rs...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// AssignPermissions assigns permissions to a group
// TODO @shipperizer see if it's worth using only one between Permission and ofga.Tuple
func (s *Service) AssignPermissions(ctx context.Context, ID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.AssignPermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]ofga.Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *ofga.NewTuple(s.buildGroupMember(ctx, ID), p.Relation, p.Object))
	}

	err := s.ofga.WriteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// RemovePermissions removes permissions from a group
// TODO @shipperizer see if it's worth using only one between Permission and ofga.Tuple
func (s *Service) RemovePermissions(ctx context.Context, ID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.RemovePermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]ofga.Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *ofga.NewTuple(s.buildGroupMember(ctx, ID), p.Relation, p.Object))
	}

	err := s.ofga.DeleteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// DeleteGroup deletes a group and all the related tuples
func (s *Service) DeleteGroup(ctx context.Context, ID string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.DeleteGroup")
	defer span.End()

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	// https://go.dev/ref/spec#Send_statements
	// A send on an unbuffered channel can proceed if a receiver is ready.
	// A send on a buffered channel can proceed if there is room in the buffer
	results := make(chan *pool.Result[any], len(s.permissionTypes()))
	wg := sync.WaitGroup{}
	wg.Add(len(s.permissionTypes()))

	// TODO @shipperizer use a background operator
	for _, t := range s.permissionTypes() {
		s.wpool.Submit(
			s.removePermissionsFunc(ctx, ID, t),
			results,
			&wg,
		)
	}

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)

	return s.ofga.DeleteTuples(ctx, *ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("group:%s", ID)))
}

// ListIdentities returns all the identities (users for now) assigned to a group
func (s *Service) ListIdentities(ctx context.Context, ID, continuationToken string) ([]string, string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.ListIdentities")
	defer span.End()

	r, err := s.ofga.ReadTuples(ctx, "", MEMBER_RELATION, fmt.Sprintf("group:%s", ID), continuationToken)

	if err != nil {
		s.logger.Error(err.Error())
		return nil, "", err
	}

	identities := make([]string, 0)

	for _, t := range r.GetTuples() {
		// TODO @shipperizer the user: bit will have to change when or if we use the identity type, this will be tricky
		// best way right now might be to verify if a user is also an identity (no idea how though)
		// at the moment an identity cannot be a member of a group, only a user
		if strings.HasPrefix(t.Key.User, "user:") {
			identities = append(identities, t.Key.User)
		}
	}

	return identities, r.GetContinuationToken(), nil
}

// AssignIdentities assigns identities to a group, right now using the type user which is disconnected
// form the identity type
func (s *Service) AssignIdentities(ctx context.Context, ID string, identities ...string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.AssignIdentities")
	defer span.End()

	ids := make([]ofga.Tuple, 0)

	for _, identity := range identities {
		// TODO @shipperizer swap user for identity if/when model changes
		ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:%s", identity), MEMBER_RELATION, fmt.Sprintf("group:%s", ID)))
	}

	err := s.ofga.WriteTuples(ctx, ids...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// RemoveIdentities removes identities from a group
func (s *Service) RemoveIdentities(ctx context.Context, ID string, identities ...string) error {
	ctx, span := s.tracer.Start(ctx, "groups.Service.RemoveIdentities")
	defer span.End()

	ids := make([]ofga.Tuple, 0)

	for _, identity := range identities {
		// TODO @shipperizer swap user for identity if/when model changes
		ids = append(ids, *ofga.NewTuple(fmt.Sprintf("user:%s", identity), MEMBER_RELATION, fmt.Sprintf("group:%s", ID)))
	}

	err := s.ofga.DeleteTuples(ctx, ids...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// TODO @shipperizer make this more scalable by pushing to a channel and using goroutine pool
// potentially create a background operator that can pipe results to an on demand channel and works off a
// set amount of goroutines
func (s *Service) listPermissionsByType(ctx context.Context, ID, pType, continuationToken string) ([]string, string, error) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.listPermissionsByType")
	defer span.End()

	r, err := s.ofga.ReadTuples(ctx, s.buildGroupMember(ctx, ID), "", fmt.Sprintf("%s:", pType), continuationToken)

	if err != nil {
		s.logger.Error(err.Error())
		return nil, "", err
	}

	permissions := make([]string, 0)

	for _, t := range r.GetTuples() {
		// if relation doesn't start with can_ it means it's not a permission (see #assignee)
		if !strings.HasPrefix(t.Key.Relation, "can_") {
			continue
		}

		permissions = append(permissions, authorization.NewUrn(t.Key.Relation, t.Key.Object).ID())
	}

	return permissions, r.GetContinuationToken(), nil
}

func (s *Service) removePermissionsByType(ctx context.Context, ID, pType string) {
	ctx, span := s.tracer.Start(ctx, "groups.Service.removePermissionsByType")
	defer span.End()

	cToken := ""
	memberRelation := s.buildGroupMember(ctx, ID)
	permissions := make([]ofga.Tuple, 0)
	for {
		r, err := s.ofga.ReadTuples(ctx, memberRelation, "", fmt.Sprintf("%s:", pType), cToken)

		if err != nil {
			s.logger.Errorf("error when retrieving tuples for %s %s", memberRelation, pType)
			return
		}

		for _, t := range r.Tuples {
			permissions = append(permissions, *ofga.NewTuple(memberRelation, t.Key.Relation, t.Key.Object))
		}

		// if there are more pages, keep going with the loop
		if cToken = r.ContinuationToken; cToken != "" {
			continue
		}

		// TODO @shipperizer understand if better breaking at every cycle or reverting if clause
		break
	}

	if err := s.ofga.DeleteTuples(ctx, permissions...); err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Service) listPermissionsFunc(ctx context.Context, groupID, ofgaType, cToken string) func() any {
	return func() any {
		p, token, err := s.listPermissionsByType(
			ctx,
			groupID,
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

func (s *Service) removePermissionsFunc(ctx context.Context, groupID, ofgaType string) func() {
	return func() {
		s.removePermissionsByType(ctx, groupID, ofgaType)
	}
}

func (s *Service) permissionTypes() []string {
	return []string{"group", "role", "identity", "scheme", "provider", "client"}
}

// NewService returns the implementtation of the business logic for the groups API
func NewService(ofga OpenFGAClientInterface, wpool pool.WorkerPoolInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.ofga = ofga

	// TODO @shipperizer make this an input
	//s.wpool = pool.NewWorkerPool(100, tracer, monitor, logger)
	s.wpool = wpool

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
