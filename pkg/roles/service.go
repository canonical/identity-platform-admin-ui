// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package roles

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
)

type listPermissionsResult struct {
	permissions []string
	token       string
	ofgaType    string
	err         error
}

// Service contains the business logic to deal with roles on the Admin UI OpenFGA model
type Service struct {
	ofga ofga.OpenFGAClientInterface
	repo RoleRepositoryInterface

	wpool pool.WorkerPoolInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// ListRoles returns all the roles a specific user can see (using "can_view" OpenFGA relation)
func (s *Service) ListRoles(ctx context.Context, userID string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListRoles")
	defer span.End()

	roles, err := s.repo.ListRoles(ctx, userID, 0, 200)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return roles, nil
}

// ListRoleGroups returns all the groups associated to a specific role
func (s *Service) ListRoleGroups(ctx context.Context, roleName string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.ListRoleGroups")
	defer span.End()

	groups, err := s.ofga.ListUsers(ctx, "group#member", authorization.ASSIGNEE_RELATION, authorization.RoleForTuple(roleName))
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return groups, nil
}

// GetRole returns the specified role using the ID argument, userID is used to validate the visibility by the user
// making the call
func (s *Service) GetRole(ctx context.Context, userID, roleName string) (*Role, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.GetRole")
	defer span.End()

	role, err := s.repo.FindRoleByNameAndOwner(ctx, roleName, userID)

	if errors.Is(err, storage.ErrNotFound) {
		s.logger.Debugf("role %s for owner %s not found", roleName, userID)
		return nil, nil
	}

	if err != nil {
		err = fmt.Errorf("unable to get role %s for owner %s, %w", roleName, userID, err)
		s.logger.Error(err.Error())
		return nil, err
	}

	return role, nil
}

// CreateRole creates a role and associates it with the userID passed as argument
// an extra tuple is created to estabilish the "privileged" relation for admin users
func (s *Service) CreateRole(ctx context.Context, userID, roleName string) (*Role, error) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.CreateRole")
	defer span.End()

	createdRole, tx, err := s.repo.CreateRoleTx(ctx, roleName, userID)
	if err != nil {
		err = fmt.Errorf("unable to create role %s for user %s, %w", roleName, userID, err)
		s.logger.Error(err.Error())
		return nil, err
	}

	role := fmt.Sprintf("role:%s", roleName)
	user := fmt.Sprintf("user:%s", userID)

	tuples := []ofga.Tuple{
		*ofga.NewTuple(user, authorization.ASSIGNEE_RELATION, role),
		*ofga.NewTuple(user, authorization.CAN_DELETE, role),
	}

	err = s.ofga.WriteTuples(ctx, tuples...)

	if err != nil {
		rollbackErr := tx.Rollback()

		err = errors.Join(err, rollbackErr)
		s.logger.Error(err.Error())
		return nil, err
	}

	// if commit fails, we rollback db transaction and the newly created tuples on OpenFGA
	if err = tx.Commit(); err != nil {
		rollbackErr := tx.Rollback()
		deleteTuplesErr := s.ofga.DeleteTuples(ctx, tuples...)

		err = errors.Join(err, rollbackErr, deleteTuplesErr)
		s.logger.Error(err.Error())
		return nil, err
	}
	s.logger.Security().AdminAction(
		userID,
		"created",
		"role",
		roleName,
		logging.WithContext(ctx),
	)

	return createdRole, nil
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
		ps = append(ps, *ofga.NewTuple(s.getRoleAssigneeUser(ID), p.Relation, p.Object))
	}

	err := s.ofga.WriteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	s.logger.Security().AdminAction(
		authentication.PrincipalFromContext(ctx).Identifier(),
		"assigned permissions to",
		"role",
		ID,
		logging.WithContext(ctx),
	)

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
		ps = append(ps, *ofga.NewTuple(s.getRoleAssigneeUser(ID), p.Relation, p.Object))
	}

	err := s.ofga.DeleteTuples(ctx, ps...)

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	s.logger.Security().AdminAction(
		authentication.PrincipalFromContext(ctx).Identifier(),
		"removed permissions from",
		"role",
		ID,
		logging.WithContext(ctx),
		logging.WithLabel("permissions", fmt.Sprintf("%+v", permissions)),
	)

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
	errorList := make([]error, 0)

	for r := range results {
		s.logger.Info(results)
		v := r.Value.(listPermissionsResult)
		permissions = append(permissions, v.permissions...)
		tMap[v.ofgaType] = v.token

		if v.err != nil {
			errorList = append(errorList, v.err)
		}
	}

	if len(errorList) == 0 {
		return permissions, tMap, nil
	}

	eMsg := ""

	for n, e := range errorList {
		s.logger.Errorf(e.Error())
		eMsg = fmt.Sprintf("%v - %s\n", n, e.Error())
	}

	return permissions, tMap, errors.New(eMsg)
}

// DeleteRole returns all the permissions associated to a specific role
func (s *Service) DeleteRole(ctx context.Context, roleName string) error {
	ctx, span := s.tracer.Start(ctx, "roles.Service.DeleteRole")
	defer span.End()

	// TODO: @barco,@shipperizer use DeleteRoleTx when we actually check errors from pool submitted jobs
	_, err := s.repo.DeleteRoleByName(ctx, roleName)
	if err != nil {
		err = fmt.Errorf("unable to delete role %s, %w", roleName, err)
		s.logger.Error(err.Error())
		return err
	}

	// keep it a buffered channel, if set to unbuffered we would need a goroutine
	// to consume from it before pushing to it
	permissionTypes := s.permissionTypes()
	directRelations := s.directRelations()

	jobs := len(permissionTypes) + len(directRelations)

	results := make(chan *pool.Result[any], jobs)
	wg := sync.WaitGroup{}
	wg.Add(jobs)

	// TODO @shipperizer use a background operator
	for _, t := range permissionTypes {
		s.wpool.Submit(
			s.removePermissionsFunc(ctx, roleName, t),
			results,
			&wg,
		)
	}

	for _, t := range directRelations {
		s.wpool.Submit(
			s.removeDirectAssociationsFunc(ctx, roleName, t),
			results,
			&wg,
		)
	}

	// wait for tasks to finish
	wg.Wait()

	// close result channel
	close(results)
	s.logger.Security().AdminAction(
		authentication.PrincipalFromContext(ctx).Identifier(),
		"delete",
		"role",
		roleName,
		logging.WithContext(ctx),
	)

	// TODO: @barco collect errors from results chan and return composite error or single summing up
	return nil
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
		permissions = append(permissions, authorization.NewURN(t.Key.Relation, t.Key.Object).ID())
	}

	return permissions, r.GetContinuationToken(), nil
}

func (s *Service) removePermissionsByType(ctx context.Context, ID, pType string) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.removePermissionsByType")
	defer span.End()

	cToken := ""
	assigneeRelation := s.getRoleAssigneeUser(ID)
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

func (s *Service) removeDirectAssociations(ctx context.Context, ID, relation string) {
	ctx, span := s.tracer.Start(ctx, "roles.Service.removeDirectAssociations")
	defer span.End()

	cToken := ""
	directs := make([]ofga.Tuple, 0)
	for {
		r, err := s.ofga.ReadTuples(ctx, "", relation, fmt.Sprintf("role:%s", ID), cToken)

		if err != nil {
			s.logger.Errorf("error when retrieving tuples for %s role, %s relation", relation, ID)
			return
		}

		for _, t := range r.Tuples {
			directs = append(directs, *ofga.NewTuple(t.Key.User, t.Key.Relation, t.Key.Object))
		}

		// if there are more pages, keep going with the loop
		if cToken = r.ContinuationToken; cToken != "" {
			continue
		}

		break
	}

	if len(directs) == 0 {
		return
	}

	if err := s.ofga.DeleteTuples(ctx, directs...); err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Service) listPermissionsFunc(ctx context.Context, roleID, ofgaType, cToken string) func() any {
	return func() any {
		p, token, err := s.listPermissionsByType(
			ctx,
			s.getRoleAssigneeUser(roleID),
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

func (s *Service) removeDirectAssociationsFunc(ctx context.Context, roleID, relation string) func() {
	return func() {
		s.removeDirectAssociations(ctx, roleID, relation)
	}
}

func (s *Service) permissionTypes() []string {
	return []string{"role", "group", "identity", "scheme", "provider", "client"}
}

func (s *Service) directRelations() []string {
	return []string{"privileged", "assignee", "can_create", "can_delete", "can_edit", "can_view"}
}

func (s *Service) getRoleAssigneeUser(roleID string) string {
	return fmt.Sprintf("role:%s#%s", roleID, authorization.ASSIGNEE_RELATION)
}

// NewService returns the implementtation of the business logic for the roles API
func NewService(ofga ofga.OpenFGAClientInterface, repo RoleRepositoryInterface, wpool pool.WorkerPoolInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.ofga = ofga
	s.repo = repo
	s.wpool = wpool

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}

type V1Service struct {
	core *Service
}

// TODO @shipperizer make sure Authenticator is implemented
// ListRoles returns a page of Role objects of at least `size` elements if available.
func (s *V1Service) ListRoles(ctx context.Context, params *resources.GetRolesParams) (*resources.PaginatedResponse[resources.Role], error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.ListRoles")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)

	if principal == nil {
		return nil, v1.NewAuthorizationError("unauthorized")
	}
	roles, err := s.core.ListRoles(ctx, principal.Identifier())

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	r := new(resources.PaginatedResponse[resources.Role])
	r.Data = make([]resources.Role, 0)
	r.Meta = resources.ResponseMeta{Size: len(roles)}

	for _, role := range roles {
		r.Data = append(r.Data, resources.Role{Id: &role, Name: role})
	}

	return r, nil
}

// CreateRole creates a single Role.
func (s *V1Service) CreateRole(ctx context.Context, role *resources.Role) (*resources.Role, error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.CreateRole")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)

	if principal == nil {
		return nil, v1.NewAuthorizationError("unauthorized")
	}
	r, err := s.core.CreateRole(ctx, principal.Identifier(), role.Name)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	role.Id = &r.ID

	// TODO @shipperizer this is quite a change from v0, happy to drop it
	if role.Entitlements == nil || len(*role.Entitlements) == 0 {
		return role, nil
	}

	permissions := make([]Permission, 0)

	for _, e := range *role.Entitlements {
		permissions = append(
			permissions,
			Permission{
				Relation: *e.Entitlement,
				Object:   *e.Resource,
			},
		)
	}

	if err := s.core.AssignPermissions(ctx, r.ID, permissions...); err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}
	// ###################################
	return role, nil
}

// GetRole returns a single Role.
func (s *V1Service) GetRole(ctx context.Context, roleId string) (*resources.Role, error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.GetRole")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)

	if principal == nil {
		return nil, v1.NewAuthorizationError("unauthorized")
	}
	r, err := s.core.GetRole(ctx, principal.Identifier(), roleId)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	if r == nil {
		return nil, v1.NewNotFoundError("role not found")
	}

	role := new(resources.Role)

	role.Id = &r.ID
	role.Name = r.Name

	return role, nil
}

// UpdateRole updates a Role.
func (s *V1Service) UpdateRole(ctx context.Context, role *resources.Role) (*resources.Role, error) {
	_, span := s.core.tracer.Start(ctx, "roles.V1Service.UpdateRole")
	defer span.End()

	return nil, v1.NewNotImplementedError("endpoint not implemented")
}

func (s *V1Service) DeleteRole(ctx context.Context, roleId string) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.DeleteRoleByName")
	defer span.End()

	if err := s.core.DeleteRole(ctx, roleId); err != nil {
		return false, v1.NewUnknownError(err.Error())
	}

	return true, nil
}

// GetRoleEntitlements returns a page of Entitlements for Role `roleId`.
func (s *V1Service) GetRoleEntitlements(ctx context.Context, roleId string, params *resources.GetRolesItemEntitlementsParams) (*resources.PaginatedResponse[resources.EntityEntitlement], error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.GetRoleEntitlements")
	defer span.End()

	paginator := types.NewTokenPaginator(s.core.tracer, s.core.logger)

	if err := paginator.LoadFromString(ctx, *params.NextToken); err != nil {
		s.core.logger.Error(err)
	}

	permissions, pageTokens, err := s.core.ListPermissions(ctx, roleId, paginator.GetAllTokens(ctx))

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	paginator.SetTokens(ctx, pageTokens)
	metaParam, err := paginator.PaginationHeader(ctx)
	if err != nil {
		s.core.logger.Errorf("error producing pagination meta param: %s", err)
		metaParam = ""
	}

	r := new(resources.PaginatedResponse[resources.EntityEntitlement])
	r.Meta = resources.ResponseMeta{Size: len(permissions)}
	r.Data = make([]resources.EntityEntitlement, 0)
	r.Next.PageToken = &metaParam

	for _, permission := range permissions {
		p := authorization.NewURNFromURLParam(permission)
		entity := strings.SplitN(p.Object(), ":", 2)
		r.Data = append(
			r.Data,
			resources.EntityEntitlement{
				Entitlement: p.Relation(),
				EntityType:  entity[0],
				EntityId:    entity[1],
			},
		)
	}

	return r, nil
}

// PatchRoleEntitlements performs addition or removal of an Entitlement to/from a Role.
func (s *V1Service) PatchRoleEntitlements(ctx context.Context, roleId string, entitlementPatches []resources.RoleEntitlementsPatchItem) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "roles.V1Service.PatchRoleEntitlements")
	defer span.End()

	additions := make([]Permission, 0)
	removals := make([]Permission, 0)
	for _, p := range entitlementPatches {
		permission := Permission{
			Relation: p.Entitlement.Entitlement,
			Object:   fmt.Sprintf("%s:%s", p.Entitlement.EntityType, p.Entitlement.EntityId),
		}

		if p.Op == "add" {
			additions = append(additions, permission)
		} else if p.Op == "remove" {
			removals = append(removals, permission)
		}
	}

	if len(additions) > 0 {
		err := s.core.AssignPermissions(ctx, roleId, additions...)

		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	if len(removals) > 0 {
		err := s.core.RemovePermissions(ctx, roleId, removals...)
		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	return true, nil
}

func NewV1Service(svc *Service) *V1Service {
	s := new(V1Service)

	s.core = svc

	return s
}
