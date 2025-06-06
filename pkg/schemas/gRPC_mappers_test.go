// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	kClient "github.com/ory/kratos-client-go"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
	"testing"
)

//go:generate mockgen -build_flags=--mod=mod -package schemas -destination ./mock_logger.go -source=../../internal/logging/interfaces.go

func TestGrpcPbMapper_FromIdentitySchemaContainerModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sc := map[string]any{
		"$id":     "https://schemas.ory.sh/presets/kratos/identity.email.schema.json",
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": map[string]any{
			"traits": map[string]any{
				"properties": map[string]any{
					"email": map[string]any{
						"format": "email",
					},
				},
			},
		},
	}
	ap := map[string]any{}
	schemaContainers := []kClient.IdentitySchemaContainer{
		{
			Id:                   strPtr("id1"),
			Schema:               sc,
			AdditionalProperties: ap,
		},
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name  string
		input []kClient.IdentitySchemaContainer
		want  []*v0Schemas.Schema
	}{
		{
			name:  "Successful mapping from kratos identity schema model to Schema",
			input: schemaContainers,
			want: []*v0Schemas.Schema{
				{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
		},
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.FromIdentitySchemaContainerModel(test.input)
			if err != nil {
				t.Errorf("FromIdentitySchemaContainerModel() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FromIdentitySchemaContainerModel() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcPbMapper_ToIdentitySchemaContainerModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sc := map[string]any{
		"$id":     "https://schemas.ory.sh/presets/kratos/identity.email.schema.json",
		"$schema": "http://json-schema.org/draft-04/schema#",
		"properties": map[string]any{
			"traits": map[string]any{
				"properties": map[string]any{
					"email": map[string]any{
						"format": "email",
					},
				},
			},
		},
	}
	ap := map[string]any{}
	schemaContainers := kClient.IdentitySchemaContainer{
		Id:                   strPtr("id1"),
		Schema:               sc,
		AdditionalProperties: nil,
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name  string
		input *v0Schemas.Schema
		want  *kClient.IdentitySchemaContainer
	}{
		{
			name: "Successful mapping from Schema to kratos identity schema model",
			input: &v0Schemas.Schema{
				Id:                   "id1",
				Schema:               schema,
				AdditionalProperties: additionalProperties,
			},
			want: &schemaContainers,
		},
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.ToIdentitySchemaContainerModel(test.input)
			if err != nil {
				t.Errorf("ToIdentitySchemaContainerModel() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ToIdentitySchemaContainerModel() got = %v, want %v", got, test.want)
			}
		})
	}
}
