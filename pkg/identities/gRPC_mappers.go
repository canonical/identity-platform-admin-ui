// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"encoding/json"
	"fmt"
	"time"

	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
	kClient "github.com/ory/kratos-client-go"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
)

// Mapping methods names prefixes
// kratos model to protobuf model => `[Ff]rom*`
// protobuf model to kratos model => [Tt]o*

type GrpcPbMapper struct {
	logger logging.LoggerInterface
}

func (g *GrpcPbMapper) FromIdentitiesModel(identities []kClient.Identity) ([]*v0Identities.Identity, error) {
	if identities == nil {
		return nil, nil
	}

	ret := make([]*v0Identities.Identity, 0, len(identities))
	for _, identity := range identities {
		mappedIdentity, err := g.fromIdentity(identity)
		if err != nil {
			return nil, err
		}

		ret = append(ret, mappedIdentity)
	}

	return ret, nil
}

func (g *GrpcPbMapper) fromIdentity(identity kClient.Identity) (*v0Identities.Identity, error) {

	organizationId := g.fromNullableString(identity.OrganizationId)
	state := identity.GetState()
	credentials := g.fromIdentityCredentials(identity.Credentials)

	metadataAdmin, err := g.fromStruct(identity.MetadataAdmin)
	if err != nil {
		return nil, err
	}

	metadataPublic, err := g.fromStruct(identity.MetadataPublic)
	if err != nil {
		return nil, err
	}

	recoveryAddresses, err := g.fromRecoveryAddresses(identity.RecoveryAddresses)
	if err != nil {
		return nil, err
	}

	verifiableAddresses, err := g.fromVerifiableAddresses(identity.VerifiableAddresses)
	if err != nil {
		return nil, err
	}

	traits, err := g.fromStruct(identity.Traits)
	if err != nil {
		return nil, err
	}

	additionalProperties, err := g.fromStringAnyMap(identity.AdditionalProperties)
	if err != nil {
		return nil, err
	}

	return &v0Identities.Identity{
		CreatedAt:            g.fromTime(identity.CreatedAt),
		Credentials:          credentials,
		Id:                   identity.GetId(),
		MetadataAdmin:        metadataAdmin,
		MetadataPublic:       metadataPublic,
		OrganizationId:       organizationId,
		RecoveryAddresses:    recoveryAddresses,
		SchemaId:             identity.GetSchemaId(),
		SchemaUrl:            identity.GetSchemaUrl(),
		State:                &state,
		StateChangedAt:       g.fromTime(identity.StateChangedAt),
		Traits:               traits,
		UpdatedAt:            g.fromTime(identity.UpdatedAt),
		VerifiableAddresses:  verifiableAddresses,
		AdditionalProperties: additionalProperties,
	}, nil
}

func (g *GrpcPbMapper) fromTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

func (g *GrpcPbMapper) fromNullableString(organizationId kClient.NullableString) *v0Identities.NullableString {
	return &v0Identities.NullableString{
		Value: organizationId.Get(),
		IsSet: organizationId.IsSet(),
	}
}

func (g *GrpcPbMapper) fromIdentityCredentials(credentials *map[string]kClient.IdentityCredentials) map[string]*v0Identities.IdentityCredentials {
	if credentials == nil {
		return nil
	}

	creds := make(map[string]*v0Identities.IdentityCredentials, len(*credentials))
	for k, v := range *credentials {
		config, _ := g.fromStringAnyMap(v.Config)
		additionalProps, _ := g.fromStringAnyMap(v.AdditionalProperties)

		creds[k] = &v0Identities.IdentityCredentials{
			Config:               config,
			CreatedAt:            g.fromTime(v.CreatedAt),
			Identifiers:          v.Identifiers,
			Type:                 v.Type,
			UpdatedAt:            g.fromTime(v.UpdatedAt),
			Version:              v.Version,
			AdditionalProperties: additionalProps,
		}
	}

	return creds
}

func (g *GrpcPbMapper) fromStringAnyMap(m map[string]interface{}) (map[string]*structpb.Value, error) {
	if m == nil {
		return nil, nil
	}

	ret := make(map[string]*structpb.Value, len(m))

	for key, value := range m {
		mappedValue, err := structpb.NewValue(value)
		if err != nil {
			g.logger.Errorf("error converting stringValuePtrMap key %s, value %v : %v", key, value, err)
			return nil, fmt.Errorf("error converting stringValuePtrMap key %s, value %v : %v", key, value, err)
		}

		ret[key] = mappedValue
	}

	return ret, nil
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

func (g *GrpcPbMapper) fromRecoveryAddresses(addresses []kClient.RecoveryIdentityAddress) ([]*v0Identities.RecoveryIdentityAddress, error) {
	if addresses == nil {
		return nil, nil
	}

	ret := make([]*v0Identities.RecoveryIdentityAddress, 0, len(addresses))

	for _, address := range addresses {
		ptrMap, err := g.fromStringAnyMap(address.AdditionalProperties)
		if err != nil {
			g.logger.Errorf("error converting recovery addresses additionalProperties %v : %v", address, err)
			return nil, fmt.Errorf("error converting recovery addresses additionalProperties %v : %v", address, err)
		}

		ret = append(
			ret,
			&v0Identities.RecoveryIdentityAddress{
				CreatedAt:            g.fromTime(address.CreatedAt),
				Id:                   address.Id,
				UpdatedAt:            g.fromTime(address.UpdatedAt),
				Value:                address.Value,
				Via:                  address.Via,
				AdditionalProperties: ptrMap,
			},
		)
	}

	return ret, nil
}

func (g *GrpcPbMapper) fromVerifiableAddresses(addresses []kClient.VerifiableIdentityAddress) ([]*v0Identities.VerifiableIdentityAddress, error) {
	if addresses == nil {
		return nil, nil
	}

	ret := make([]*v0Identities.VerifiableIdentityAddress, 0, len(addresses))
	for _, address := range addresses {
		ptrMap, err := g.fromStringAnyMap(address.AdditionalProperties)
		if err != nil {
			g.logger.Errorf("error converting verifiable addresses additionalProperties %v : %v", address, err)
			return nil, fmt.Errorf("error converting verifiable addresses additionalProperties %v : %v", address, err)
		}

		ret = append(
			ret,
			&v0Identities.VerifiableIdentityAddress{
				CreatedAt:            g.fromTime(address.CreatedAt),
				Id:                   address.Id,
				Status:               address.Status,
				UpdatedAt:            g.fromTime(address.UpdatedAt),
				Value:                address.Value,
				Verified:             address.Verified,
				VerifiedAt:           g.fromTime(address.VerifiedAt),
				Via:                  address.Via,
				AdditionalProperties: ptrMap,
			},
		)
	}

	return ret, nil
}

func (g *GrpcPbMapper) ToCreateIdentityModel(body *v0Identities.CreateIdentityBody) (*kClient.CreateIdentityBody, error) {
	if body == nil {
		return nil, nil
	}

	var credentials *kClient.IdentityWithCredentials = nil
	if body.Credentials != nil {
		if err := g.mapStringInterfaceMap(body.GetCredentials().AsMap(), &credentials); err != nil {
			return nil, err
		}
	}

	var metadataAdmin interface{} = nil
	if body.MetadataAdmin != nil {
		asMap := body.GetMetadataAdmin().AsMap()
		if len(asMap) != 0 {
			metadataAdmin = asMap
		}
	}
	var metadataPublic interface{} = nil
	if body.GetMetadataPublic() != nil {
		asMap := body.GetMetadataPublic().AsMap()
		if len(asMap) != 0 {
			metadataPublic = body.GetMetadataPublic().AsMap()
		}
	}

	var recoveryAddresses []kClient.RecoveryIdentityAddress = nil
	if body.RecoveryAddresses != nil {
		recoveryAddresses = make([]kClient.RecoveryIdentityAddress, 0, len(body.RecoveryAddresses))

		for _, address := range body.RecoveryAddresses {
			recoveryAddresses = append(recoveryAddresses, g.toRecoveryAddress(address))
		}
	}

	var verifiableAddresses []kClient.VerifiableIdentityAddress = nil
	if body.VerifiableAddresses != nil {
		verifiableAddresses = make([]kClient.VerifiableIdentityAddress, 0, len(body.VerifiableAddresses))

		for _, address := range body.VerifiableAddresses {
			verifiableAddresses = append(verifiableAddresses, g.toVerifiableAddress(address))
		}
	}

	var additionalProperties map[string]interface{} = nil
	if body.GetAdditionalProperties() != nil {
		asMap := body.AdditionalProperties.AsMap()
		if len(asMap) != 0 {
			additionalProperties = body.AdditionalProperties.AsMap()
		}
	}

	var traits map[string]interface{} = nil
	if body.GetTraits() != nil {
		asMap := body.GetTraits().AsMap()
		if len(asMap) != 0 {
			traits = body.GetTraits().AsMap()
		}
	}

	ret := &kClient.CreateIdentityBody{
		Credentials:          credentials,
		MetadataAdmin:        metadataAdmin,
		MetadataPublic:       metadataPublic,
		RecoveryAddresses:    recoveryAddresses,
		SchemaId:             body.SchemaId,
		State:                &body.State,
		Traits:               traits,
		VerifiableAddresses:  verifiableAddresses,
		AdditionalProperties: additionalProperties,
	}

	return ret, nil
}

func (g *GrpcPbMapper) toRecoveryAddress(address *v0Identities.RecoveryIdentityAddress) kClient.RecoveryIdentityAddress {
	ret := kClient.RecoveryIdentityAddress{
		Id:                   address.GetId(),
		Value:                address.GetValue(),
		Via:                  address.GetVia(),
		AdditionalProperties: g.toStringInterfaceMapFromStringValueMap(address.AdditionalProperties),
	}

	if address.GetCreatedAt() != nil {
		asTime := address.GetCreatedAt().AsTime()
		ret.CreatedAt = &asTime
	}

	if address.GetUpdatedAt() != nil {
		asTime := address.GetUpdatedAt().AsTime()
		ret.UpdatedAt = &asTime
	}

	return ret
}

func (g *GrpcPbMapper) toVerifiableAddress(address *v0Identities.VerifiableIdentityAddress) kClient.VerifiableIdentityAddress {
	ret := kClient.VerifiableIdentityAddress{
		Id:                   address.Id,
		Status:               address.GetStatus(),
		Value:                address.GetValue(),
		Verified:             address.GetVerified(),
		Via:                  address.GetVia(),
		AdditionalProperties: g.toStringInterfaceMapFromStringValueMap(address.AdditionalProperties),
	}

	if address.GetCreatedAt() != nil {
		asTime := address.GetCreatedAt().AsTime()
		ret.CreatedAt = &asTime
	}

	if address.GetUpdatedAt() != nil {
		asTime := address.GetUpdatedAt().AsTime()
		ret.UpdatedAt = &asTime
	}

	if address.GetVerifiedAt() != nil {
		asTime := address.GetVerifiedAt().AsTime()
		ret.VerifiedAt = &asTime
	}

	return ret
}

func (g *GrpcPbMapper) toStringInterfaceMapFromStringValueMap(m map[string]*structpb.Value) map[string]interface{} {
	if m == nil {
		return nil
	}

	ret := make(map[string]interface{}, len(m))
	for k, v := range m {
		ret[k] = v.AsInterface()
	}

	return ret
}

func (g *GrpcPbMapper) mapStringInterfaceMap(m map[string]any, model any) error {
	if m == nil || len(m) == 0 {
		return nil
	}

	bytesBody, err := json.Marshal(m)
	if err != nil {
		g.logger.Errorf("error marshalling credentials %v : %v", m, err)
		return fmt.Errorf("error marshalling credentials %v : %v", m, err)
	}

	err = json.Unmarshal(bytesBody, &model)
	if err != nil {
		g.logger.Errorf("error unmarshalling credentials %v : %v", bytesBody, err)
		return fmt.Errorf("error unmarshalling credentials %v : %v", bytesBody, err)
	}

	return nil
}

func (g *GrpcPbMapper) ToUpdateIdentityModel(body *v0Identities.UpdateIdentityBody) (*kClient.UpdateIdentityBody, error) {
	if body == nil {
		return nil, nil
	}

	var credentials *kClient.IdentityWithCredentials = nil
	if body.Credentials != nil {
		if err := g.mapStringInterfaceMap(body.GetCredentials().AsMap(), &credentials); err != nil {
			return nil, err
		}
	}

	var metadataAdmin interface{} = nil
	if body.MetadataAdmin != nil {
		asMap := body.GetMetadataAdmin().AsMap()
		if len(asMap) != 0 {
			metadataAdmin = body.GetMetadataAdmin().AsMap()
		}
	}
	var metadataPublic interface{} = nil
	if body.GetMetadataPublic() != nil {
		asMap := body.GetMetadataPublic().AsMap()
		if len(asMap) != 0 {
			metadataPublic = body.GetMetadataPublic().AsMap()
		}
	}

	var traits map[string]interface{} = nil
	if body.GetTraits() != nil {
		asMap := body.GetTraits().AsMap()
		if len(asMap) != 0 {
			traits = body.GetTraits().AsMap()
		}
	}

	var additionalProperties map[string]interface{} = nil
	if body.GetAdditionalProperties() != nil {
		asMap := body.GetAdditionalProperties().AsMap()
		if len(asMap) != 0 {
			additionalProperties = body.AdditionalProperties.AsMap()
		}
	}

	ret := &kClient.UpdateIdentityBody{
		Credentials:          credentials,
		MetadataAdmin:        metadataAdmin,
		MetadataPublic:       metadataPublic,
		SchemaId:             body.GetSchemaId(),
		State:                body.GetState(),
		Traits:               traits,
		AdditionalProperties: additionalProperties,
	}

	return ret, nil
}

func NewGrpcMapper(logger logging.LoggerInterface) *GrpcPbMapper {
	return &GrpcPbMapper{
		logger: logger,
	}
}
