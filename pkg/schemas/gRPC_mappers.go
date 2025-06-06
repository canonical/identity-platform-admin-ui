// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"encoding/json"
	"fmt"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	kClient "github.com/ory/kratos-client-go"
	"google.golang.org/protobuf/types/known/structpb"
)

type GrpcPbMapper struct {
	logger logging.LoggerInterface
}

func (g *GrpcPbMapper) FromIdentitySchemaContainerModel(schemas []kClient.IdentitySchemaContainer) ([]*v0Schemas.Schema, error) {
	if schemas == nil {
		return nil, nil
	}

	ret := make([]*v0Schemas.Schema, 0, len(schemas))
	for _, schema := range schemas {
		mappedSchema, err := g.fromIdentitySchemaContainer(schema)
		if err != nil {
			return nil, err
		}

		ret = append(ret, mappedSchema)
	}

	return ret, nil
}

func (g *GrpcPbMapper) ToIdentitySchemaContainerModel(body *v0Schemas.Schema) (*kClient.IdentitySchemaContainer, error) {
	if body == nil {
		return nil, nil
	}

	var schema map[string]interface{}
	if body.GetSchema() != nil {
		asMap := body.GetSchema().AsMap()
		if len(asMap) != 0 {
			schema = asMap
		}
	}

	var additionalProperties map[string]interface{}
	if body.GetAdditionalProperties() != nil {
		asMap := body.GetAdditionalProperties().AsMap()
		if len(asMap) != 0 {
			additionalProperties = asMap
		}
	}

	return &kClient.IdentitySchemaContainer{
		Id:                   &body.Id,
		Schema:               schema,
		AdditionalProperties: additionalProperties,
	}, nil
}

func (g *GrpcPbMapper) fromIdentitySchemaContainer(schemaContainer kClient.IdentitySchemaContainer) (*v0Schemas.Schema, error) {
	schema, err := g.toStruct(schemaContainer.Schema)
	if err != nil {
		return nil, err
	}

	additionalProperties, err := g.toStruct(schemaContainer.AdditionalProperties)
	if err != nil {
		return nil, err
	}

	return &v0Schemas.Schema{
		Id:                   *schemaContainer.Id,
		Schema:               schema,
		AdditionalProperties: additionalProperties,
	}, nil
}

func (g *GrpcPbMapper) toStruct(value interface{}) (*structpb.Struct, error) {
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

func NewGrpcMapper(logger logging.LoggerInterface) *GrpcPbMapper {
	return &GrpcPbMapper{
		logger: logger,
	}
}
