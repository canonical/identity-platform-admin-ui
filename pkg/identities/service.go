// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

type Service struct {
	kratos kClient.IdentityAPI
	authz  AuthorizerInterface

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
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.ListIdentities")
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
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.GetIdentity")
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
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.CreateIdentity")
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

func (s *Service) UpdateIdentity(ctx context.Context, ID string, bodyID *kClient.UpdateIdentityBody) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.UpdateIdentity")
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
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.DeleteIdentity")
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

func NewService(kratos kClient.IdentityAPI, authz AuthorizerInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.kratos = kratos
	s.authz = authz

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
