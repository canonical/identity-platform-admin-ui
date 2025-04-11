// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
	kClient "github.com/ory/kratos-client-go"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_logger.go -source=../../internal/logging/interfaces.go

func TestGrpcPbMapper_FromIdentitiesModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTime := time.Date(2025, 4, 25, 12, 12, 12, 0, time.UTC).UTC()
	mockMappedTime := timestamppb.New(mockTime)

	mockMetadata := make(map[string]interface{})
	nestedMap := make(map[string]interface{})
	nestedMap["nested-key"] = "test"
	mockMetadata["test-key"] = nestedMap

	mockState := "test-state"

	mockCreds := make(map[string]kClient.IdentityCredentials)
	mockOidcCreds := kClient.NewIdentityCredentials()
	mockOidcCreds.SetCreatedAt(mockTime)
	mockOidcCreds.SetUpdatedAt(mockTime)
	mockOidcCreds.SetConfig(mockMetadata)
	mockOidcCreds.SetVersion(1)
	mockOidcCreds.SetIdentifiers([]string{"mock-identifier"})
	mockOidcCreds.SetType("oidc")

	mockCreds["oidc"] = *mockOidcCreds

	mockMappedConfig := make(map[string]*structpb.Value, 1)
	mockMappedStruct, _ := structpb.NewStruct(map[string]any{"nested-key": "test"})
	mockMappedConfig["test-key"] = structpb.NewStructValue(mockMappedStruct)

	mockMappedMetadata, _ := structpb.NewStruct(mockMetadata)

	var mockVersion int64 = 1
	mockTestValueString := "oidc"
	mockMappedCreds := make(map[string]*v0Identities.IdentityCredentials, 1)

	mockMappedOidcCreds := v0Identities.IdentityCredentials{
		Config:      mockMappedConfig,
		CreatedAt:   mockMappedTime,
		Identifiers: []string{"mock-identifier"},
		Type:        &mockTestValueString,
		UpdatedAt:   mockMappedTime,
		Version:     &mockVersion,
	}
	mockMappedCreds["oidc"] = &mockMappedOidcCreds

	mockId := "mock-id"

	mockRecoveryAddresses := make([]kClient.RecoveryIdentityAddress, 0, 1)
	mockRecoveryAddresses = append(
		mockRecoveryAddresses,
		kClient.RecoveryIdentityAddress{
			CreatedAt: &mockTime,
			Id:        mockId,
			UpdatedAt: &mockTime,
			Value:     "mock-value",
			Via:       "mock-via",
		},
	)

	mockMappedRecoveryAddresses := make([]*v0Identities.RecoveryIdentityAddress, 0, 1)
	mockMappedRecoveryAddresses = append(
		mockMappedRecoveryAddresses,
		&v0Identities.RecoveryIdentityAddress{
			CreatedAt: mockMappedTime,
			Id:        mockId,
			UpdatedAt: mockMappedTime,
			Value:     "mock-value",
			Via:       "mock-via",
		},
	)

	mockVerifiableAddresses := make([]kClient.VerifiableIdentityAddress, 0, 1)
	mockVerifiableAddresses = append(
		mockVerifiableAddresses,
		kClient.VerifiableIdentityAddress{
			CreatedAt:  &mockTime,
			Id:         &mockId,
			Status:     "mock-status",
			UpdatedAt:  &mockTime,
			Value:      "mock-value",
			Verified:   true,
			VerifiedAt: &mockTime,
			Via:        "mock-via",
		},
	)

	mockMappedVerifiableAddresses := make([]*v0Identities.VerifiableIdentityAddress, 0, 1)
	mockMappedVerifiableAddresses = append(
		mockMappedVerifiableAddresses,
		&v0Identities.VerifiableIdentityAddress{
			CreatedAt:  mockMappedTime,
			Id:         &mockId,
			Status:     "mock-status",
			UpdatedAt:  mockMappedTime,
			Value:      "mock-value",
			Verified:   true,
			VerifiedAt: mockMappedTime,
			Via:        "mock-via",
		},
	)

	tests := []struct {
		name       string
		identities []kClient.Identity
		want       []*v0Identities.Identity
	}{
		{
			name: "Successful mapping with full structs",
			identities: []kClient.Identity{
				{
					CreatedAt:            &mockTime,
					Credentials:          &mockCreds,
					Id:                   "test-id",
					MetadataAdmin:        mockMetadata,
					MetadataPublic:       mockMetadata,
					OrganizationId:       *kClient.NewNullableString(&mockState),
					RecoveryAddresses:    mockRecoveryAddresses,
					SchemaId:             "test-schema-id",
					SchemaUrl:            "test://schema-url",
					State:                &mockState,
					StateChangedAt:       &mockTime,
					Traits:               mockMetadata,
					UpdatedAt:            &mockTime,
					VerifiableAddresses:  mockVerifiableAddresses,
					AdditionalProperties: mockMetadata,
				},
			},
			want: []*v0Identities.Identity{
				{
					CreatedAt:      mockMappedTime,
					Credentials:    mockMappedCreds,
					Id:             "test-id",
					MetadataAdmin:  mockMappedMetadata,
					MetadataPublic: mockMappedMetadata,
					OrganizationId: &v0Identities.NullableString{
						Value: &mockState,
						IsSet: true,
					},
					RecoveryAddresses:    mockMappedRecoveryAddresses,
					SchemaId:             "test-schema-id",
					SchemaUrl:            "test://schema-url",
					State:                &mockState,
					StateChangedAt:       mockMappedTime,
					Traits:               mockMappedMetadata,
					UpdatedAt:            mockMappedTime,
					VerifiableAddresses:  mockMappedVerifiableAddresses,
					AdditionalProperties: mockMappedConfig,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			g := NewGrpcMapper(mockLogger)

			got, err := g.FromIdentitiesModel(tt.identities)
			if err != nil {
				t.Errorf("FromIdentitiesModel() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromIdentitiesModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcPbMapper_ToCreateIdentityModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"

	mappedMockCreds := &kClient.IdentityWithCredentials{
		Oidc: &kClient.IdentityWithCredentialsOidc{
			Config: &kClient.IdentityWithCredentialsOidcConfig{
				Providers: []kClient.IdentityWithCredentialsOidcConfigProvider{
					{
						Provider: "mock-provider",
						Subject:  "mock-subject",
					},
				},
			},
		},
	}

	marshal, _ := json.Marshal(mappedMockCreds)

	var mappedMockCredsMap map[string]any
	_ = json.Unmarshal(marshal, &mappedMockCredsMap)

	mockCreds, _ := structpb.NewStruct(mappedMockCredsMap)

	tests := []struct {
		name string
		body *v0Identities.CreateIdentityBody
		want *kClient.CreateIdentityBody
	}{
		{
			name: "Successful mapping",
			body: &v0Identities.CreateIdentityBody{
				Credentials: mockCreds,
				SchemaId:    "mock-schema-id",
				State:       mockState,
			},
			want: &kClient.CreateIdentityBody{
				Credentials: mappedMockCreds,
				SchemaId:    "mock-schema-id",
				State:       &mockState,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			g := NewGrpcMapper(mockLogger)

			got, err := g.ToCreateIdentityModel(tt.body)

			if err != nil {
				t.Errorf("ToCreateIdentityModel() error = %v", err)
				return
			}

			if got == nil || tt.want == nil {
				if got != tt.want {
					t.Errorf("got is %v, want is %v", got, tt.want)
				}
				return
			}

			// normalize map for testing purposes (due to underlying implementation leveraging json unmarshalling
			got.Credentials.Oidc.Config.Providers[0].AdditionalProperties = nil

			if !reflect.DeepEqual(got.Credentials.Oidc.Config.Providers[0], tt.want.Credentials.Oidc.Config.Providers[0]) {
				t.Errorf("Credentials mismatch:\ngot  = %#v\nwant = %#v", got.Credentials.Oidc.Config.Providers[0], tt.want.Credentials.Oidc.Config.Providers[0])
			}

			if got.SchemaId != tt.want.SchemaId {
				t.Errorf("SchemaId mismatch:\ngot  = %s\nwant = %s", got.SchemaId, tt.want.SchemaId)
			}

			if got.State == nil || tt.want.State == nil {
				if got.State != tt.want.State {
					t.Errorf("State pointer mismatch:\ngot  = %v\nwant = %v", got.State, tt.want.State)
				}
			} else if *got.State != *tt.want.State {
				t.Errorf("State value mismatch:\ngot  = %s\nwant = %s", *got.State, *tt.want.State)
			}
		})
	}
}

func TestGrpcPbMapper_ToUpdateIdentityModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mappedMockCreds := &kClient.IdentityWithCredentials{
		Oidc: &kClient.IdentityWithCredentialsOidc{
			Config: &kClient.IdentityWithCredentialsOidcConfig{
				Providers: []kClient.IdentityWithCredentialsOidcConfigProvider{
					{
						Provider: "mock-provider",
						Subject:  "mock-subject",
					},
				},
			},
		},
	}

	marshal, _ := json.Marshal(mappedMockCreds)

	var mappedMockCredsMap map[string]any
	_ = json.Unmarshal(marshal, &mappedMockCredsMap)

	mockCreds, _ := structpb.NewStruct(mappedMockCredsMap)

	tests := []struct {
		name string
		body *v0Identities.UpdateIdentityBody
		want *kClient.UpdateIdentityBody
	}{
		{
			name: "Successful mapping with full structs",
			body: &v0Identities.UpdateIdentityBody{
				Credentials: mockCreds,
				SchemaId:    "mock-schema-id",
				State:       "mock-state",
			},
			want: &kClient.UpdateIdentityBody{
				Credentials: mappedMockCreds,
				SchemaId:    "mock-schema-id",
				State:       "mock-state",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			g := NewGrpcMapper(mockLogger)

			got, err := g.ToUpdateIdentityModel(tt.body)
			if err != nil {
				t.Errorf("ToUpdateIdentityModel() error = %v", err)
				return
			}

			// normalize map for testing purposes (due to underlying implementation leveraging json unmarshalling
			got.Credentials.Oidc.Config.Providers[0].AdditionalProperties = nil

			if !reflect.DeepEqual(got.Credentials.Oidc.Config.Providers[0], tt.want.Credentials.Oidc.Config.Providers[0]) {
				t.Errorf("Credentials mismatch:\ngot  = %#v\nwant = %#v", got.Credentials.Oidc.Config.Providers[0], tt.want.Credentials.Oidc.Config.Providers[0])
			}

			if got.SchemaId != tt.want.SchemaId {
				t.Errorf("SchemaId mismatch:\ngot  = %s\nwant = %s", got.SchemaId, tt.want.SchemaId)
			}

			if got.State != tt.want.State {
				t.Errorf("State value mismatch:\ngot  = %s\nwant = %s", got.State, tt.want.State)
			}
		})
	}
}
