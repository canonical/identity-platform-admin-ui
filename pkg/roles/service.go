// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

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
)

const (
	ASSIGNEE_RELATION = "assignee"
)

type Service struct {
	ofga OpenFGAClientInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

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

// ListRoleGroups does rely on the /read endpoint which allows for pagination via the token
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

func (s *Service) ListPermissions(ctx context.Context, ID string, continuationTokens map[string]string) ([]string, map[string]string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListPermissions")
	defer span.End()

	permissionsMap := sync.Map{}
	tokensMap := sync.Map{}

	var wg sync.WaitGroup

	wg.Add(len(s.permissionTypes()))

	// TODO @shipperizer use a background operator
	for _, t := range s.permissionTypes() {
		go func(pType string) {
			defer wg.Done()
			p, t, err := s.listPermissionsByType(ctx, fmt.Sprintf("role:%s#%s", ID, ASSIGNEE_RELATION), pType, continuationTokens[pType])

			permissionsMap.Store(pType, p)
			tokensMap.Store(pType, t)

			// TODO @shipperizer handle errors better
			// chain them and return at the end of the function
			if err != nil {
				s.logger.Error(err)
			}
		}(t)
	}

	wg.Wait()

	permissions := make([]string, 0)
	tokens := make(map[string]string)

	permissionsMap.Range(
		func(key any, value any) bool {
			permissions = append(permissions, value.([]string)...)

			return true
		},
	)

	tokensMap.Range(
		func(key any, value any) bool {
			tokens[key.(string)] = value.(string)

			return true
		},
	)

	// TODO @shipperizer right now the function fails silently, chain errors from the goroutines
	// and return
	return permissions, tokens, nil
}

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

func (s *Service) DeleteRole(ctx context.Context, ID string) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.DeleteRole")
	defer span.End()

	var wg sync.WaitGroup

	wg.Add(len(s.permissionTypes()))

	// TODO @shipperizer use a background operator
	for _, t := range s.permissionTypes() {
		go func(pType string) {
			defer wg.Done()
			s.removePermissionsByType(ctx, ID, pType)
		}(t)
	}

	wg.Wait()

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

	if err := s.ofga.DeleteTuples(ctx, permissions...); err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Service) permissionTypes() []string {
	return []string{"role", "group", "identity", "scheme", "provider", "client"}
}

func NewService(ofga OpenFGAClientInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.ofga = ofga
	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
