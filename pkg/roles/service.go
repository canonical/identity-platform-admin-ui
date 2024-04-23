// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

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
	ASSIGNEE_RELATION = "assignee"
)

type listPermissionsResult struct {
	permissions []string
	token       string
	ofgaType    string
	err         error
}

// Service contains the business logic to deal with roles on the Admin UI OpenFGA model
type Service struct {
	ofga OpenFGAClientInterface

	wpool pool.WorkerPoolInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// ListRoles returns all the roles a specific user can see (using "can_view" OpenFGA relation)
func (s *Service) ListRoles(ctx context.Context, userID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListRoles")
	defer span.End()

	roles, err := s.ofga.ListObjects(ctx, fmt.Sprintf("user:%s", userID), "can_view", "role")

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return roles, nil
}

// ListRoleGroups returns all the groups associated to a specific role
// method relies on the /read endpoint which allows for pagination via the token
// unfortunately we are not able to distinguish between types assigned on the OpenFGA side,
// so we'll have to filter here based on the user, this leads to unrealiable object counts
// TODO @shipperizer a more complex pagination system can be implemented by keeping track of the
// latest index in the current "page" and encode it in the pagination token header returned to
// the UI
func (s *Service) ListRoleGroups(ctx context.Context, ID, continuationToken string) ([]string, string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListRoleGroups")
	defer span.End()

	r, err := s.ofga.ReadTuples(ctx, "", ASSIGNEE_RELATION, fmt.Sprintf("role:%s", ID), continuationToken)

	if err != nil {
		s.logger.Error(err.Error())
		return nil, "", err
	}

	groups := make([]string, 0)

	for _, t := range r.GetTuples() {
		if strings.HasPrefix(t.Key.User, "group:") {
			groups = append(groups, t.Key.User)
		}
	}

	return groups, r.GetContinuationToken(), nil
}

// GetRole returns the specified role using the ID argument, userID is used to validate the visibility by the user
// making the call
func (s *Service) GetRole(ctx context.Context, userID, ID string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.GetRole")
	defer span.End()

	exists, err := s.ofga.Check(ctx, fmt.Sprintf("user:%s", userID), "can_view", fmt.Sprintf("role:%s", ID))

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

// CreateRole creates a role and associates it with the userID passed as argument
// an extra tuple is created to estabilish the "privileged" relatin for admin users
func (s *Service) CreateRole(ctx context.Context, userID, ID string) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.CreateRole")
	defer span.End()

	// TODO @shipperizer will we need also the can_view?
	// does creating a role mean that you are the owner, therefore u get all the permissions on it?
	// right now assumption is only admins will be able to do this
	// potentially changing the model to say
	// `define can_view: [user, user:*, role#assignee, group#member] or assignee or admin from privileged`
	// might sort the problem

	// TODO @shipperizer offload to privileged creator object
	err := s.ofga.WriteTuples(
		ctx,
		*ofga.NewTuple(fmt.Sprintf("user:%s", userID), ASSIGNEE_RELATION, fmt.Sprintf("role:%s", ID)),
		*ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("role:%s", ID)),
	)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// AssignPermissions assigns permissions to a role
// TODO @shipperizer see if it's worth using only one between Permission and ofga.Tuple
func (s *Service) AssignPermissions(ctx context.Context, ID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.AssignPermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]ofga.Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *ofga.NewTuple(fmt.Sprintf("role:%s#%s", ID, ASSIGNEE_RELATION), p.Relation, p.Object))
	}

	err := s.ofga.WriteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// RemovePermissions removes permissions from a role
// TODO @shipperizer see if it's worth using only one between Permission and ofga.Tuple
func (s *Service) RemovePermissions(ctx context.Context, ID string, permissions ...Permission) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.RemovePermissions")
	defer span.End()

	// preemptive check to verify if all permissions to be assigned are accessible by the user
	// needs to happen separately

	ps := make([]ofga.Tuple, 0)

	for _, p := range permissions {
		ps = append(ps, *ofga.NewTuple(fmt.Sprintf("role:%s#%s", ID, ASSIGNEE_RELATION), p.Relation, p.Object))
	}

	err := s.ofga.DeleteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

// ListPermissions returns all the permissions associated to a specific role
func (s *Service) ListPermissions(ctx context.Context, ID string, continuationTokens map[string]string) ([]string, map[string]string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListPermissions")
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
		eMsg = fmt.Sprintf("%v - %s\n", n, e.Error())
	}

	return permissions, tMap, fmt.Errorf(eMsg)
}

// DeleteRole returns all the permissions associated to a specific role
func (s *Service) DeleteRole(ctx context.Context, ID string) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.DeleteRole")
	defer span.End()

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	// https://go.dev/ref/spec#Send_statements
	// A send on an unbuffered channel can proceed if a receiver is ready.
	// A send on a buffered channel can proceed if there is room in the buffer
	jobs := len(s.permissionTypes()) + 1

	results := make(chan *pool.Result[any], jobs)
	wg := sync.WaitGroup{}
	// number of types + 1 for assignees job
	wg.Add(jobs)

	// TODO @shipperizer use a background operator
	for _, t := range s.permissionTypes() {
		s.wpool.Submit(
			s.removePermissionsFunc(ctx, ID, t),
			results,
			&wg,
		)
	}

	s.wpool.Submit(
		s.removeAssigneesFunc(ctx, ID),
		results,
		&wg,
	)

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)

	return s.ofga.DeleteTuples(ctx, *ofga.NewTuple(authorization.ADMIN_PRIVILEGE, "privileged", fmt.Sprintf("role:%s", ID)))
}

// TODO @shipperizer make this more scalable by pushing to a channel and using goroutine pool
// potentially create a background operator that can pipe results to an on demand channel and works off a
// set amount of goroutines
func (s *Service) listPermissionsByType(ctx context.Context, roleIDAssignee, pType, continuationToken string) ([]string, string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.listPermissionsByType")
	defer span.End()

	r, err := s.ofga.ReadTuples(ctx, roleIDAssignee, "", fmt.Sprintf("%s:", pType), continuationToken)

	if err != nil {
		s.logger.Error(err.Error())
		return nil, "", err
	}

	permissions := make([]string, 0)

	for _, t := range r.GetTuples() {
		permissions = append(permissions, authorization.NewUrn(t.Key.Relation, t.Key.Object).ID())
	}

	return permissions, r.GetContinuationToken(), nil
}

func (s *Service) removePermissionsByType(ctx context.Context, ID, pType string) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.removePermissionsByType")
	defer span.End()

	cToken := ""
	assigneeRelation := fmt.Sprintf("role:%s#%s", ID, ASSIGNEE_RELATION)
	permissions := make([]ofga.Tuple, 0)
	for {
		r, err := s.ofga.ReadTuples(ctx, assigneeRelation, "", fmt.Sprintf("%s:", pType), cToken)

		if err != nil {
			s.logger.Errorf("error when retrieving tuples for %s %s", assigneeRelation, pType)
			return
		}

		for _, t := range r.Tuples {
			permissions = append(permissions, *ofga.NewTuple(assigneeRelation, t.Key.Relation, t.Key.Object))
		}

		// if there are more pages, keep going with the loop
		if cToken = r.ContinuationToken; cToken != "" {
			continue
		}

		// TODO @shipperizer understand if better breaking at every cycle or reverting if clause
		break
	}

	if len(permissions) == 0 {
		return
	}

	if err := s.ofga.DeleteTuples(ctx, permissions...); err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Service) removeAssignees(ctx context.Context, ID string) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.removeAssignees")
	defer span.End()

	cToken := ""
	assignees := make([]ofga.Tuple, 0)
	for {
		r, err := s.ofga.ReadTuples(ctx, "", ASSIGNEE_RELATION, fmt.Sprintf("role:%s", ID), cToken)

		if err != nil {
			s.logger.Errorf("error when retrieving tuples for %s role:%s", ASSIGNEE_RELATION, ID)
			return
		}

		for _, t := range r.Tuples {
			assignees = append(assignees, *ofga.NewTuple(t.Key.User, t.Key.Relation, t.Key.Object))
		}

		// if there are more pages, keep going with the loop
		if cToken = r.ContinuationToken; cToken != "" {
			continue
		}

		// TODO @shipperizer understand if better breaking at every cycle or reverting if clause
		break
	}

	if len(assignees) == 0 {
		return
	}

	if err := s.ofga.DeleteTuples(ctx, assignees...); err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Service) listPermissionsFunc(ctx context.Context, roleID, ofgaType, cToken string) func() any {
	return func() any {
		p, token, err := s.listPermissionsByType(
			ctx,
			fmt.Sprintf("role:%s#%s", roleID, ASSIGNEE_RELATION),
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

func (s *Service) removePermissionsFunc(ctx context.Context, roleID, ofgaType string) func() {
	return func() {
		s.removePermissionsByType(ctx, roleID, ofgaType)
	}
}

func (s *Service) removeAssigneesFunc(ctx context.Context, roleID string) func() {
	return func() {
		s.removeAssignees(ctx, roleID)
	}
}

func (s *Service) permissionTypes() []string {
	return []string{"role", "group", "identity", "scheme", "provider", "client"}
}

// NewService returns the implementtation of the business logic for the roles API
func NewService(ofga OpenFGAClientInterface, wpool pool.WorkerPoolInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.ofga = ofga
	s.wpool = wpool

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
