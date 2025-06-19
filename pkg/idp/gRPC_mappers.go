// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"encoding/json"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
)

type GrpcPbMapper struct {
	logger logging.LoggerInterface
}

func (g *GrpcPbMapper) FromConfigurations(configurations []*Configuration) ([]*v0Idps.Idp, error) {
	if configurations == nil {
		return nil, nil
	}

	ret := make([]*v0Idps.Idp, 0, len(configurations))
	for _, configuration := range configurations {
		idp, err := g.fromConfiguration(configuration)
		if err != nil {
			return nil, err
		}

		ret = append(ret, idp)
	}

	return ret, nil
}

func (g *GrpcPbMapper) ToCreateIdpBody(idp *v0Idps.CreateIdpBody) (*Configuration, error) {
	if idp == nil {
		return nil, nil
	}

	rc, err := json.Marshal(idp.GetRequestedClaims())
	if err != nil {
		return nil, err
	}

	return &Configuration{
		ID:              idp.GetId(),
		Provider:        idp.GetProvider(),
		Label:           idp.GetLabel(),
		ClientID:        idp.GetClientId(),
		ClientSecret:    idp.GetClientSecret(),
		IssuerURL:       idp.GetIssuerUrl(),
		AuthURL:         idp.GetAuthUrl(),
		TokenURL:        idp.GetTokenUrl(),
		Tenant:          idp.GetMicrosoftTenant(),
		SubjectSource:   idp.GetSubjectSource(),
		TeamId:          idp.GetAppleTeamId(),
		PrivateKeyId:    idp.GetApplePrivateKeyId(),
		PrivateKey:      idp.GetApplePrivateKey(),
		Scope:           idp.GetScope(),
		Mapper:          idp.GetMapperUrl(),
		RequestedClaims: rc,
	}, nil
}

func (g *GrpcPbMapper) ToUpdateIdpBody(idp *v0Idps.UpdateIdpBody) (*Configuration, error) {
	if idp == nil {
		return nil, nil
	}

	rc, err := json.Marshal(idp.GetRequestedClaims())
	if err != nil {
		return nil, err
	}

	return &Configuration{
		ID:              idp.GetId(),
		Provider:        idp.GetProvider(),
		Label:           idp.GetLabel(),
		ClientID:        idp.GetClientId(),
		ClientSecret:    idp.GetClientSecret(),
		IssuerURL:       idp.GetIssuerUrl(),
		AuthURL:         idp.GetAuthUrl(),
		TokenURL:        idp.GetTokenUrl(),
		Tenant:          idp.GetMicrosoftTenant(),
		SubjectSource:   idp.GetSubjectSource(),
		TeamId:          idp.GetAppleTeamId(),
		PrivateKeyId:    idp.GetApplePrivateKeyId(),
		PrivateKey:      idp.GetApplePrivateKey(),
		Scope:           idp.GetScope(),
		Mapper:          idp.GetMapperUrl(),
		RequestedClaims: rc,
	}, nil
}

func (g *GrpcPbMapper) fromConfiguration(configuration *Configuration) (*v0Idps.Idp, error) {
	var requestedClaims string

	if configuration.RequestedClaims != nil {
		err := json.Unmarshal(configuration.RequestedClaims, &requestedClaims)
		if err != nil {
			return nil, err
		}
	}

	return &v0Idps.Idp{
		Id:                configuration.ID,
		Provider:          configuration.Provider,
		Label:             &configuration.Label,
		ClientId:          configuration.ClientID,
		ClientSecret:      &configuration.ClientSecret,
		IssuerUrl:         &configuration.IssuerURL,
		AuthUrl:           &configuration.AuthURL,
		TokenUrl:          &configuration.TokenURL,
		MicrosoftTenant:   &configuration.Tenant,
		SubjectSource:     &configuration.SubjectSource,
		AppleTeamId:       &configuration.TeamId,
		ApplePrivateKeyId: &configuration.PrivateKeyId,
		ApplePrivateKey:   &configuration.PrivateKey,
		Scope:             configuration.Scope,
		MapperUrl:         &configuration.Mapper,
		RequestedClaims:   &requestedClaims,
	}, nil
}

func NewGrpcMapper(logger logging.LoggerInterface) *GrpcPbMapper {
	return &GrpcPbMapper{
		logger: logger,
	}
}
