// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"encoding/json"
	"fmt"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	hClient "github.com/ory/hydra-client-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"time"
)

type GrpcPbMapper struct {
	logger logging.LoggerInterface
}

func (g *GrpcPbMapper) FromOAuth2Clients(clients []hClient.OAuth2Client) ([]*v0Clients.Client, error) {
	if clients == nil {
		return nil, nil
	}

	ret := make([]*v0Clients.Client, 0, len(clients))
	for _, client := range clients {
		mappedClient, err := g.fromOAuth2Client(&client)
		if err != nil {
			return nil, err
		}

		ret = append(ret, mappedClient)
	}

	return ret, nil
}

func (g *GrpcPbMapper) ToOAuth2Client(client *v0Clients.Client) (*hClient.OAuth2Client, error) {
	if client == nil {
		return nil, nil
	}

	var createdAt *time.Time = nil
	if client.GetCreatedAt() != nil {
		asTime := client.GetCreatedAt().AsTime()
		createdAt = &asTime
	}

	var updatedAt *time.Time = nil
	if client.GetUpdatedAt() != nil {
		asTime := client.GetUpdatedAt().AsTime()
		updatedAt = &asTime
	}

	var jwks interface{} = nil
	if client.GetJwks() != nil {
		asMap := client.GetJwks().AsMap()
		if len(asMap) != 0 {
			jwks = asMap
		}
	}

	var metadata interface{} = nil
	if client.GetMetadata() != nil {
		asMap := client.GetMetadata().AsMap()
		if len(asMap) != 0 {
			metadata = asMap
		}
	}

	tokenEndpointAuthMethod := "client_secret_basic"

	return &hClient.OAuth2Client{
		AccessTokenStrategy: client.AccessTokenStrategy,
		AllowedCorsOrigins:  client.GetAllowedCorsOrigins(),
		Audience:            client.GetAudience(),
		AuthorizationCodeGrantAccessTokenLifespan:  client.AuthorizationCodeGrantAccessTokenLifespan,
		AuthorizationCodeGrantIdTokenLifespan:      client.AuthorizationCodeGrantIdTokenLifespan,
		AuthorizationCodeGrantRefreshTokenLifespan: client.AuthorizationCodeGrantRefreshTokenLifespan,
		BackchannelLogoutSessionRequired:           client.BackchannelLogoutSessionRequired,
		BackchannelLogoutUri:                       client.BackchannelLogoutUri,
		ClientCredentialsGrantAccessTokenLifespan:  client.ClientCredentialsGrantAccessTokenLifespan,
		ClientId:                              client.ClientId,
		ClientName:                            client.ClientName,
		ClientSecret:                          client.ClientSecret,
		ClientSecretExpiresAt:                 client.ClientSecretExpiresAt,
		ClientUri:                             client.ClientUri,
		Contacts:                              client.GetContacts(),
		CreatedAt:                             createdAt,
		FrontchannelLogoutSessionRequired:     client.FrontchannelLogoutSessionRequired,
		FrontchannelLogoutUri:                 client.FrontchannelLogoutUri,
		GrantTypes:                            client.GetGrantTypes(),
		ImplicitGrantAccessTokenLifespan:      client.ImplicitGrantAccessTokenLifespan,
		ImplicitGrantIdTokenLifespan:          client.ImplicitGrantIdTokenLifespan,
		Jwks:                                  jwks,
		JwksUri:                               client.JwksUri,
		JwtBearerGrantAccessTokenLifespan:     client.JwtBearerGrantAccessTokenLifespan,
		LogoUri:                               client.LogoUri,
		Metadata:                              metadata,
		Owner:                                 client.Owner,
		PolicyUri:                             client.PolicyUri,
		PostLogoutRedirectUris:                client.GetPostLogoutRedirectUris(),
		RedirectUris:                          client.GetRedirectUris(),
		RefreshTokenGrantAccessTokenLifespan:  client.RefreshTokenGrantAccessTokenLifespan,
		RefreshTokenGrantIdTokenLifespan:      client.RefreshTokenGrantIdTokenLifespan,
		RefreshTokenGrantRefreshTokenLifespan: client.RefreshTokenGrantRefreshTokenLifespan,
		RegistrationAccessToken:               client.RegistrationAccessToken,
		RegistrationClientUri:                 client.RegistrationClientUri,
		RequestObjectSigningAlg:               client.RequestObjectSigningAlg,
		RequestUris:                           client.GetRequestUris(),
		ResponseTypes:                         client.GetResponseTypes(),
		Scope:                                 client.Scope,
		SectorIdentifierUri:                   client.SectorIdentifierUri,
		SkipConsent:                           client.SkipConsent,
		SubjectType:                           client.SubjectType,
		TokenEndpointAuthMethod:               &tokenEndpointAuthMethod,
		TokenEndpointAuthSigningAlg:           client.TokenEndpointAuthSigningAlg,
		TosUri:                                client.TosUri,
		UpdatedAt:                             updatedAt,
		UserinfoSignedResponseAlg:             client.UserinfoSignedResponseAlg,
	}, nil
}

func (g *GrpcPbMapper) fromOAuth2Client(client *hClient.OAuth2Client) (*v0Clients.Client, error) {
	jwks, err := g.fromStruct(client.GetJwks())
	if err != nil {
		return nil, err
	}

	metadata, err := g.fromStruct(client.GetMetadata())
	if err != nil {
		return nil, err
	}

	return &v0Clients.Client{
		AccessTokenStrategy: client.AccessTokenStrategy,
		AllowedCorsOrigins:  client.AllowedCorsOrigins,
		Audience:            client.GetAudience(),
		AuthorizationCodeGrantAccessTokenLifespan:  client.AuthorizationCodeGrantAccessTokenLifespan,
		AuthorizationCodeGrantIdTokenLifespan:      client.AuthorizationCodeGrantIdTokenLifespan,
		AuthorizationCodeGrantRefreshTokenLifespan: client.AuthorizationCodeGrantRefreshTokenLifespan,
		BackchannelLogoutSessionRequired:           client.BackchannelLogoutSessionRequired,
		BackchannelLogoutUri:                       client.BackchannelLogoutUri,
		ClientCredentialsGrantAccessTokenLifespan:  client.ClientCredentialsGrantAccessTokenLifespan,
		ClientId:                              client.ClientId,
		ClientName:                            client.ClientName,
		ClientSecret:                          client.ClientSecret,
		ClientSecretExpiresAt:                 client.ClientSecretExpiresAt,
		ClientUri:                             client.ClientUri,
		Contacts:                              client.GetContacts(),
		CreatedAt:                             g.fromTime(client.CreatedAt),
		FrontchannelLogoutSessionRequired:     client.FrontchannelLogoutSessionRequired,
		FrontchannelLogoutUri:                 client.FrontchannelLogoutUri,
		GrantTypes:                            client.GetGrantTypes(),
		ImplicitGrantAccessTokenLifespan:      client.ImplicitGrantIdTokenLifespan,
		ImplicitGrantIdTokenLifespan:          client.ImplicitGrantIdTokenLifespan,
		Jwks:                                  jwks,
		JwksUri:                               client.JwksUri,
		JwtBearerGrantAccessTokenLifespan:     client.JwtBearerGrantAccessTokenLifespan,
		LogoUri:                               client.LogoUri,
		Metadata:                              metadata,
		Owner:                                 client.Owner,
		PolicyUri:                             client.PolicyUri,
		PostLogoutRedirectUris:                client.GetPostLogoutRedirectUris(),
		RedirectUris:                          client.GetRedirectUris(),
		RefreshTokenGrantAccessTokenLifespan:  client.RefreshTokenGrantAccessTokenLifespan,
		RefreshTokenGrantIdTokenLifespan:      client.RefreshTokenGrantIdTokenLifespan,
		RefreshTokenGrantRefreshTokenLifespan: client.RefreshTokenGrantRefreshTokenLifespan,
		RegistrationAccessToken:               client.RegistrationAccessToken,
		RegistrationClientUri:                 client.RegistrationClientUri,
		RequestObjectSigningAlg:               client.RequestObjectSigningAlg,
		RequestUris:                           client.GetRequestUris(),
		ResponseTypes:                         client.GetResponseTypes(),
		Scope:                                 client.Scope,
		SectorIdentifierUri:                   client.SectorIdentifierUri,
		SkipConsent:                           client.SkipConsent,
		SubjectType:                           client.SubjectType,
		TokenEndpointAuthMethod:               client.TokenEndpointAuthMethod,
		TokenEndpointAuthSigningAlg:           client.TokenEndpointAuthSigningAlg,
		TosUri:                                client.TosUri,
		UpdatedAt:                             g.fromTime(client.UpdatedAt),
		UserinfoSignedResponseAlg:             client.UserinfoSignedResponseAlg,
	}, nil
}

func (g *GrpcPbMapper) fromTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

func (g *GrpcPbMapper) fromStruct(value interface{}) (*structpb.Struct, error) {
	if value == nil {
		return nil, nil
	}

	m, err := g.toStringAnyMap(value)
	if err != nil {
		g.logger.Errorf("error transforming any to map, value %v : %v", value, err)
		return nil, fmt.Errorf("error transforming any to map, value %v : %v", value, err)
	}

	ret, err := structpb.NewStruct(m)
	if err != nil {
		g.logger.Errorf("error creating pb Struct, value %v : %v", value, err)
		return nil, fmt.Errorf("error creating pb Struct, value %v : %v", value, err)
	}

	return ret, nil
}

func (g *GrpcPbMapper) toStringAnyMap(value interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		g.logger.Errorf("error marshalling value %v : %v", value, err)
		return nil, fmt.Errorf("error marshalling value %v : %v", value, err)
	}

	var m map[string]interface{} = nil

	err = json.Unmarshal(bytes, &m)
	if err != nil {
		g.logger.Errorf("error unmarshalling value %v : %v", value, err)
		return nil, fmt.Errorf("error unmarshalling value %v : %v", value, err)
	}

	return m, nil
}

func (g *GrpcPbMapper) codeFromHTTPStatus(status int) codes.Code {
	switch status {
	case http.StatusOK:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.Aborted
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case 499:
		return codes.Canceled
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	default:
		return codes.Unknown
	}
}

func NewGrpcMapper(logger logging.LoggerInterface) *GrpcPbMapper {
	return &GrpcPbMapper{
		logger: logger,
	}
}
