// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

// TODO @shipperizer unify this value with schemas/service.go
const (
	DEFAULT_SCHEMA           = "default.schema"
	userCreationEmailSubject = "Complete your registration"
)

type Service struct {
	kratos kClient.IdentityAPI
	authz  AuthorizerInterface
	email  mail.EmailServiceInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

type IdentityData struct {
	Identities []kClient.Identity
	Tokens     types.NavigationTokens
	Error      *kClient.GenericError
}

// TODO @shipperizer verify during integration test if this is actually the format
type KratosError struct {
	Error *kClient.GenericError `json:"error,omitempty"`
}

func (s *Service) buildListRequest(ctx context.Context, size int64, token, credID string) kClient.IdentityAPIListIdentitiesRequest {
	r := s.kratos.ListIdentities(ctx).PageToken(token).PageSize(size)

	if credID != "" {
		r = r.CredentialsIdentifier(credID)
	}

	return r
}

func (s *Service) parseError(r *http.Response) *kClient.GenericError {
	gerr := KratosError{Error: kClient.NewGenericErrorWithDefaults()}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(body, &gerr); err != nil {
		gerr.Error.SetMessage("unable to parse kratos error response")
		gerr.Error.SetCode(http.StatusInternalServerError)
	}

	return gerr.Error
}

func (s *Service) ListIdentities(ctx context.Context, size int64, token, credID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "identities.Service.ListIdentities")
	defer span.End()

	identities, rr, err := s.kratos.ListIdentitiesExecute(
		s.buildListRequest(ctx, size, token, credID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if navTokens, err := types.ParseLinkTokens(rr.Header); err != nil {
		s.logger.Warnf("failed parsing link header: %s", err)
	} else {
		data.Tokens = navTokens
	}

	data.Identities = identities

	// TODO @shipperizer check if identities is defaulting to empty slice inside kratos-client
	if data.Identities == nil {
		data.Identities = make([]kClient.Identity, 0)
	}

	return data, err
}

func (s *Service) GetIdentity(ctx context.Context, ID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "identities.Service.GetIdentity")
	defer span.End()

	identity, rr, err := s.kratos.GetIdentityExecute(
		s.kratos.GetIdentity(ctx, ID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	return data, err
}

func (s *Service) CreateIdentity(ctx context.Context, bodyID *kClient.CreateIdentityBody) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "identities.Service.CreateIdentity")
	defer span.End()

	if bodyID == nil {
		err := fmt.Errorf("no identity data passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	identity, rr, err := s.kratos.CreateIdentityExecute(
		s.kratos.CreateIdentity(ctx).CreateIdentityBody(*bodyID),
	)

	data := new(IdentityData)

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
		return data, err
	}

	s.authz.SetCreateIdentityEntitlements(ctx, identity.Id)

	return data, err
}

func (s *Service) SendUserCreationEmail(ctx context.Context, identity *kClient.Identity) error {
	ctx, span := s.tracer.Start(ctx, "identities.Service.SendUserCreationEmail")
	defer span.End()

	template, err := mail.LoadTemplate(mail.UserCreationInvite)
	if err != nil {
		return err
	}

	code, link, err := s.generateRecoveryInfo(ctx, identity.Id)
	if err != nil {
		return err
	}

	emailAddress := ""
	if e, ok := identity.Traits.(map[string]interface{})["email"]; ok {
		emailAddress = e.(string)
	}

	if emailAddress == "" {
		return fmt.Errorf("\"email\" address not found in identity traits")
	}

	userCreationInviteArgs := mail.UserCreationInviteArgs{
		Email:        emailAddress,
		InviteUrl:    link,
		RecoveryCode: code,
	}

	err = s.email.Send(ctx, emailAddress, userCreationEmailSubject, template, userCreationInviteArgs)

	return err
}

func (s *Service) generateRecoveryInfo(ctx context.Context, identityId string) (string, string, error) {
	request := kClient.CreateRecoveryCodeForIdentityBody{IdentityId: identityId}
	recoveryInfo, response, err := s.kratos.CreateRecoveryCodeForIdentity(ctx).
		CreateRecoveryCodeForIdentityBody(request).
		Execute()

	if err != nil {
		return "", "", err
	}

	if response.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("unable to create recovery code for Identity %v, status code %d", identityId, response.StatusCode)
	}

	return recoveryInfo.RecoveryCode, recoveryInfo.RecoveryLink, nil
}

func (s *Service) UpdateIdentity(ctx context.Context, ID string, bodyID *kClient.UpdateIdentityBody) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "identities.Service.UpdateIdentity")
	defer span.End()
	if ID == "" {
		err := fmt.Errorf("no identity ID passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	if bodyID == nil {
		err := fmt.Errorf("no identity body passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	identity, rr, err := s.kratos.UpdateIdentityExecute(
		s.kratos.UpdateIdentity(ctx, ID).UpdateIdentityBody(*bodyID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	return data, err
}

func (s *Service) DeleteIdentity(ctx context.Context, ID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "identities.Service.DeleteIdentity")
	defer span.End()

	rr, err := s.kratos.DeleteIdentityExecute(
		s.kratos.DeleteIdentity(ctx, ID),
	)

	data := new(IdentityData)

	data.Identities = []kClient.Identity{}
	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
		return data, err
	}

	s.authz.SetDeleteIdentityEntitlements(ctx, ID)

	return data, err
}

func NewService(kratos kClient.IdentityAPI, authz AuthorizerInterface, email mail.EmailServiceInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.kratos = kratos
	s.authz = authz
	s.email = email

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}

type V1Service struct {
	cmName      string
	cmNamespace string

	k8s   coreV1.CoreV1Interface
	store OpenFGAStoreInterface

	core *Service
}

func (s *V1Service) getDefaultSchema(ctx context.Context) (string, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.getDefaultSchema")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.core.logger.Error(err.Error())
		return "", err
	}

	ID, ok := cm.Data[DEFAULT_SCHEMA]

	if !ok {
		return "", fmt.Errorf("missing default schema")
	}

	return ID, nil
}

// ListIdentities returns a page of Identity objects of at least `size` elements if available
func (s *V1Service) ListIdentities(ctx context.Context, params *resources.GetIdentitiesParams) (*resources.PaginatedResponse[resources.Identity], error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.ListIdentities")
	defer span.End()

	size := 100
	token := ""

	if params != nil && params.Size != nil {
		size = *params.Size
	}

	if params != nil && params.NextToken != nil {
		token = *params.NextToken
	}

	// TODO @shipperizer use params.Filter to fetch credID
	ids, err := s.core.ListIdentities(ctx, int64(size), token, "")

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	r := new(resources.PaginatedResponse[resources.Identity])
	r.Data = make([]resources.Identity, 0)
	r.Meta = resources.ResponseMeta{Size: len(ids.Identities), PageToken: &token}
	r.Next = resources.Next{PageToken: &ids.Tokens.Next}
	for _, id := range ids.Identities {
		traits, ok := id.Traits.(map[string]string)

		if !ok {
			traits = make(map[string]string)
		}

		// TODO @shipperizer enhance Identity resource with Permissions and Roles on the next iteration
		// this requires calls to openfga in here unless we enhance the PrincipalContext and let that do
		// the calls
		i := resources.Identity{
			Id: &id.Id,
		}

		if email, ok := traits["email"]; ok {
			i.Email = email
		}

		fullname, ok := traits["name"]

		if !ok {
			r.Data = append(r.Data, i)
			continue
		}

		surnameIndex := strings.LastIndex(fullname, " ")

		if surnameIndex > 0 {
			name := strings.Trim(fullname[0:surnameIndex], " ")
			surname := strings.Trim(fullname[surnameIndex:], " ")

			i.FirstName = &name
			i.LastName = &surname
		}

		r.Data = append(r.Data, i)
	}

	return r, nil
}

// CreateIdentity creates a single Identity.
func (s *V1Service) CreateIdentity(ctx context.Context, identity *resources.Identity) (*resources.Identity, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.CreateIdentity")
	defer span.End()

	active := "StateActive"
	schemaId, err := s.getDefaultSchema(ctx)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	if identity == nil {
		return nil, v1.NewRequestBodyValidationError("bad identity payload")
	}

	traits := make(map[string]interface{})

	traits["email"] = identity.Email

	if identity.FirstName != nil && identity.LastName != nil {
		traits["name"] = fmt.Sprintf("%s %s", *identity.FirstName, *identity.LastName)
	}

	ids, err := s.core.CreateIdentity(ctx,
		&kClient.CreateIdentityBody{
			State:    &active,
			SchemaId: schemaId,
			// TODO @shipperizer the code below assumes each schema has name and email
			// needs to be validated as schemas might differ
			Traits: traits,
		},
	)

	// TODO @shipperizer enhance Identity resource with Permissions and Roles on the next iteration
	// this requires calls to openfga in here unless we enhance the PrincipalContext and let that do
	// the calls
	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	if len(ids.Identities) != 1 {
		return nil, v1.NewInvalidRequestError("no identity created")
	}

	return &resources.Identity{
		Email:     identity.Email,
		FirstName: identity.FirstName,
		LastName:  identity.LastName,
		Id:        &ids.Identities[0].Id,
	}, nil
}

// GetIdentity returns a single Identity.
func (s *V1Service) GetIdentity(ctx context.Context, identityId string) (*resources.Identity, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.GetIdentity")
	defer span.End()

	ids, err := s.core.GetIdentity(ctx, identityId)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	if ids.Identities == nil || len(ids.Identities) != 1 {
		return nil, v1.NewNotFoundError("identity not found")
	}

	id := ids.Identities[0]

	traits, ok := id.Traits.(map[string]string)

	if !ok {
		traits = make(map[string]string)
	}

	// TODO @shipperizer enhance Identity resource with Permissions and Roles on the next iteration
	// this requires calls to openfga in here unless we enhance the PrincipalContext and let that do
	// the calls
	i := new(resources.Identity)

	i.Id = &id.Id

	if email, ok := traits["email"]; ok {
		i.Email = email
	}

	fullname, ok := traits["name"]
	if !ok {
		return i, nil
	}

	surnameIndex := strings.LastIndex(fullname, " ")

	if surnameIndex > 0 {
		name := strings.Trim(fullname[0:surnameIndex], " ")
		surname := strings.Trim(fullname[surnameIndex:], " ")

		i.FirstName = &name
		i.LastName = &surname
	}

	return i, nil
}

// UpdateIdentity updates an Identity.
func (s *V1Service) UpdateIdentity(ctx context.Context, identity *resources.Identity) (*resources.Identity, error) {
	_, span := s.core.tracer.Start(ctx, "identities.V1Service.UpdateIdentity")
	defer span.End()

	if identity == nil {
		return nil, v1.NewRequestBodyValidationError("bad identity payload")
	}

	traits := make(map[string]interface{})

	traits["email"] = identity.Email
	if identity.FirstName != nil && identity.LastName != nil {
		traits["name"] = fmt.Sprintf("%s %s", *identity.FirstName, *identity.LastName)
	}

	body := kClient.NewUpdateIdentityBodyWithDefaults()
	body.SetTraits(traits)

	ids, err := s.core.UpdateIdentity(
		ctx,
		*identity.Id,
		// TODO @shipperizer the code below assumes each schema has name and email
		// needs to be validated as schemas might differ
		body,
	)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	if len(ids.Identities) != 1 {
		return nil, v1.NewInvalidRequestError("no identity created")
	}

	id := ids.Identities[0]

	ts, ok := id.GetTraits().(map[string]string)

	if !ok {
		ts = make(map[string]string)
	}

	// TODO @shipperizer enhance Identity resource with Permissions and Roles on the next iteration
	// this requires calls to openfga in here unless we enhance the PrincipalContext and let that do
	// the calls
	i := new(resources.Identity)

	i.Id = &id.Id

	if email, ok := ts["email"]; ok {
		i.Email = email
	}

	fullname, ok := ts["name"]
	if !ok {
		return i, nil
	}

	surnameIndex := strings.LastIndex(fullname, " ")

	if surnameIndex > 0 {
		name := strings.Trim(fullname[0:surnameIndex], " ")
		surname := strings.Trim(fullname[surnameIndex:], " ")

		i.FirstName = &name
		i.LastName = &surname
	}

	return i, nil

}

// DeleteIdentity deletes an Identity
// returns (true, nil) in case an identity was successfully delete
// return (false, error) in case something went wrong
// implementors may want to return (false, nil) for idempotency cases
func (s *V1Service) DeleteIdentity(ctx context.Context, identityId string) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.DeleteIdentity")
	defer span.End()

	if _, err := s.core.DeleteIdentity(ctx, identityId); err != nil {
		return false, v1.NewUnknownError(err.Error())
	}

	return true, nil
}

// GetIdentityGroups returns a page of Groups for identity `identityId`.
func (s *V1Service) GetIdentityGroups(ctx context.Context, identityId string, params *resources.GetIdentitiesItemGroupsParams) (*resources.PaginatedResponse[resources.Group], error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.GetIdentityGroups")
	defer span.End()

	groups, err := s.store.ListAssignedGroups(ctx, fmt.Sprintf("user:%s", identityId))
	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	r := new(resources.PaginatedResponse[resources.Group])
	r.Data = make([]resources.Group, 0)
	r.Meta = resources.ResponseMeta{Size: len(groups)}

	for _, group := range groups {
		r.Data = append(r.Data, resources.Group{Id: &group, Name: group})
	}

	return r, nil
}

// GetIdentityRoles returns a page of Roles for identity `identityId`.
func (s *V1Service) GetIdentityRoles(ctx context.Context, identityId string, params *resources.GetIdentitiesItemRolesParams) (*resources.PaginatedResponse[resources.Role], error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.GetIdentityRoles")
	defer span.End()

	roles, err := s.store.ListAssignedRoles(ctx, fmt.Sprintf("user:%s", identityId))
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

// PatchIdentityGroups performs addition or removal of Groups to/from an Identity.
func (s *V1Service) PatchIdentityGroups(ctx context.Context, identityId string, groupPatches []resources.IdentityGroupsPatchItem) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.PatchIdentityGroups")
	defer span.End()

	additions := make([]string, 0)
	removals := make([]string, 0)
	for _, p := range groupPatches {
		group := fmt.Sprintf("group:%s", p.Group)

		if p.Op == "add" {
			additions = append(additions, group)
		} else if p.Op == "remove" {
			removals = append(removals, group)
		}
	}

	if len(additions) > 0 {
		err := s.store.AssignGroups(ctx, fmt.Sprintf("user:%s", identityId), additions...)

		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	if len(removals) > 0 {
		err := s.store.UnassignGroups(ctx, fmt.Sprintf("user:%s", identityId), removals...)
		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	return true, nil
}

// PatchIdentityRoles performs addition or removal of Roles to/from an Identity.
func (s *V1Service) PatchIdentityRoles(ctx context.Context, identityId string, rolePatches []resources.IdentityRolesPatchItem) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.PatchIdentityRoles")
	defer span.End()

	additions := make([]string, 0)
	removals := make([]string, 0)
	for _, p := range rolePatches {
		role := fmt.Sprintf("role:%s", p.Role)

		if p.Op == "add" {
			additions = append(additions, role)
		} else if p.Op == "remove" {
			removals = append(removals, role)
		}
	}

	if len(additions) > 0 {
		err := s.store.AssignRoles(ctx, fmt.Sprintf("user:%s", identityId), additions...)

		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	if len(removals) > 0 {
		err := s.store.UnassignRoles(ctx, fmt.Sprintf("user:%s", identityId), removals...)
		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	return true, nil
}

// GetIdentityEntitlements returns a page of Entitlements for identity `identityId`.
func (s *V1Service) GetIdentityEntitlements(ctx context.Context, identityId string, params *resources.GetIdentitiesItemEntitlementsParams) (*resources.PaginatedResponse[resources.EntityEntitlement], error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.GetIdentityEntitlements")
	defer span.End()

	paginator := types.NewTokenPaginator(s.core.tracer, s.core.logger)

	nextToken := ""

	if params != nil && params.NextPageToken != nil {
		nextToken = *params.NextPageToken
	}

	if err := paginator.LoadFromString(ctx, nextToken); err != nil {
		s.core.logger.Error(err)
	}

	permissions, pageTokens, err := s.store.ListPermissions(ctx, fmt.Sprintf("user:%s", identityId), paginator.GetAllTokens(ctx))

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

		entity := strings.SplitN(permission.Object, ":", 2)
		r.Data = append(
			r.Data,
			resources.EntityEntitlement{
				Entitlement: permission.Relation,
				EntityType:  entity[0],
				EntityId:    entity[1],
			},
		)
	}

	return r, nil
}

// PatchIdentityEntitlements performs addition or removal of an Entitlement to/from an Identity.
func (s *V1Service) PatchIdentityEntitlements(ctx context.Context, identityId string, entitlementPatches []resources.IdentityEntitlementsPatchItem) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "identities.V1Service.PatchIdentityEntitlements")
	defer span.End()

	additions := make([]ofga.Permission, 0)
	removals := make([]ofga.Permission, 0)
	for _, p := range entitlementPatches {
		permission := ofga.Permission{
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
		err := s.store.AssignPermissions(ctx, fmt.Sprintf("user:%s", identityId), additions...)

		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	if len(removals) > 0 {
		err := s.store.UnassignPermissions(ctx, fmt.Sprintf("user:%s", identityId), removals...)
		if err != nil {
			return false, v1.NewUnknownError(err.Error())
		}
	}

	return true, nil
}

type Config struct {
	Name         string
	Namespace    string
	K8s          coreV1.CoreV1Interface
	OpenFGAStore OpenFGAStoreInterface
}

func NewV1Service(config *Config, svc *Service) *V1Service {
	s := new(V1Service)

	s.core = svc
	s.k8s = config.K8s
	s.cmName = config.Name
	s.cmNamespace = config.Namespace
	s.store = config.OpenFGAStore

	return s
}
