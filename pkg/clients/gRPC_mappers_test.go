// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	hClient "github.com/ory/hydra-client-go/v2"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"testing"
	"time"
)

//go:generate mockgen -build_flags=--mod=mod -package clients -destination ./mock_logger.go -source=../../internal/logging/interfaces.go

func TestGrpcPbMapper_FromOAuth2Clients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createdAt := time.Now()
	jwks := map[string]any{
		"jwks": "jwks",
	}
	metadata := map[string]any{
		"metadata": "metadata",
	}
	oauth2Clients := []hClient.OAuth2Client{
		{
			ClientId:     strPtr("clientId"),
			ClientName:   strPtr("clientName"),
			ClientSecret: strPtr("clientSecret"),
			CreatedAt:    &createdAt,
			GrantTypes:   []string{"authorization_code", "refresh_token"},
			Jwks:         jwks,
			Metadata:     metadata,
			RedirectUris: []string{"http://localhost"},
			Scope:        strPtr("email"),
		},
	}

	ca := timestamppb.New(createdAt)
	js, _ := structpb.NewStruct(jwks)
	md, _ := structpb.NewStruct(metadata)

	client := &v0Clients.Client{
		AccessTokenStrategy: nil,
		AllowedCorsOrigins:  nil,
		Audience:            nil,
		AuthorizationCodeGrantAccessTokenLifespan:  nil,
		AuthorizationCodeGrantIdTokenLifespan:      nil,
		AuthorizationCodeGrantRefreshTokenLifespan: nil,
		BackchannelLogoutSessionRequired:           nil,
		BackchannelLogoutUri:                       nil,
		ClientCredentialsGrantAccessTokenLifespan:  nil,
		ClientId:                              strPtr("clientId"),
		ClientName:                            strPtr("clientName"),
		ClientSecret:                          strPtr("clientSecret"),
		ClientSecretExpiresAt:                 nil,
		ClientUri:                             nil,
		Contacts:                              nil,
		CreatedAt:                             ca,
		FrontchannelLogoutSessionRequired:     nil,
		FrontchannelLogoutUri:                 nil,
		GrantTypes:                            []string{"authorization_code", "refresh_token"},
		ImplicitGrantAccessTokenLifespan:      nil,
		ImplicitGrantIdTokenLifespan:          nil,
		Jwks:                                  js,
		JwksUri:                               nil,
		JwtBearerGrantAccessTokenLifespan:     nil,
		LogoUri:                               nil,
		Metadata:                              md,
		Owner:                                 nil,
		PolicyUri:                             nil,
		PostLogoutRedirectUris:                nil,
		RedirectUris:                          []string{"http://localhost"},
		RefreshTokenGrantAccessTokenLifespan:  nil,
		RefreshTokenGrantIdTokenLifespan:      nil,
		RefreshTokenGrantRefreshTokenLifespan: nil,
		RegistrationAccessToken:               nil,
		RegistrationClientUri:                 nil,
		RequestObjectSigningAlg:               nil,
		RequestUris:                           nil,
		ResponseTypes:                         nil,
		Scope:                                 strPtr("email"),
		SectorIdentifierUri:                   nil,
		SkipConsent:                           nil,
		SubjectType:                           nil,
		TokenEndpointAuthMethod:               nil,
		TokenEndpointAuthSigningAlg:           nil,
		TosUri:                                nil,
		UpdatedAt:                             nil,
		UserinfoSignedResponseAlg:             nil,
	}

	tests := []struct {
		name  string
		input []hClient.OAuth2Client
		want  []*v0Clients.Client
	}{
		{
			name:  "Successful mapping from hydra OAuth2 client to Client",
			input: oauth2Clients,
			want:  []*v0Clients.Client{client},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.FromOAuth2Clients(test.input)
			if err != nil {
				t.Errorf("FromOAuth2Clients() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FromOAuth2Clients() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcPbMapper_ToOAuth2Client(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwks := map[string]any{
		"jwks": "jwks",
	}
	metadata := map[string]any{
		"metadata": "metadata",
	}
	oauth2Client := &hClient.OAuth2Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		Jwks:                    jwks,
		Metadata:                metadata,
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	js, _ := structpb.NewStruct(jwks)
	md, _ := structpb.NewStruct(metadata)

	client := &v0Clients.Client{
		AccessTokenStrategy: nil,
		AllowedCorsOrigins:  nil,
		Audience:            nil,
		AuthorizationCodeGrantAccessTokenLifespan:  nil,
		AuthorizationCodeGrantIdTokenLifespan:      nil,
		AuthorizationCodeGrantRefreshTokenLifespan: nil,
		BackchannelLogoutSessionRequired:           nil,
		BackchannelLogoutUri:                       nil,
		ClientCredentialsGrantAccessTokenLifespan:  nil,
		ClientId:                              strPtr("clientId"),
		ClientName:                            strPtr("clientName"),
		ClientSecret:                          strPtr("clientSecret"),
		ClientSecretExpiresAt:                 nil,
		ClientUri:                             nil,
		Contacts:                              nil,
		CreatedAt:                             nil,
		FrontchannelLogoutSessionRequired:     nil,
		FrontchannelLogoutUri:                 nil,
		GrantTypes:                            []string{"authorization_code", "refresh_token"},
		ImplicitGrantAccessTokenLifespan:      nil,
		ImplicitGrantIdTokenLifespan:          nil,
		Jwks:                                  js,
		JwksUri:                               nil,
		JwtBearerGrantAccessTokenLifespan:     nil,
		LogoUri:                               nil,
		Metadata:                              md,
		Owner:                                 nil,
		PolicyUri:                             nil,
		PostLogoutRedirectUris:                nil,
		RedirectUris:                          []string{"http://localhost"},
		RefreshTokenGrantAccessTokenLifespan:  nil,
		RefreshTokenGrantIdTokenLifespan:      nil,
		RefreshTokenGrantRefreshTokenLifespan: nil,
		RegistrationAccessToken:               nil,
		RegistrationClientUri:                 nil,
		RequestObjectSigningAlg:               nil,
		RequestUris:                           nil,
		ResponseTypes:                         nil,
		Scope:                                 strPtr("email"),
		SectorIdentifierUri:                   nil,
		SkipConsent:                           nil,
		SubjectType:                           nil,
		TokenEndpointAuthMethod:               nil,
		TokenEndpointAuthSigningAlg:           nil,
		TosUri:                                nil,
		UpdatedAt:                             nil,
		UserinfoSignedResponseAlg:             nil,
	}

	tests := []struct {
		name  string
		input *v0Clients.Client
		want  *hClient.OAuth2Client
	}{
		{
			name:  "Successful mapping from Client to hydra OAuth2 client",
			input: client,
			want:  oauth2Client,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.ToOAuth2Client(test.input)
			if err != nil {
				t.Errorf("ToOAuth2Client() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ToOAuth2Client() got = %v, want %v", got, test.want)
			}
		})
	}
}
