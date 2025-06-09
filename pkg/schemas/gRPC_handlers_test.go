// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package schemas

import (
	"context"
	"errors"
	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	v0Types "github.com/canonical/identity-platform-api/v0/http"
	v0Schemas "github.com/canonical/identity-platform-api/v0/schemas"
	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"net/http"
	"reflect"
	"testing"
)

//go:generate mockgen -build_flags=--mod=mod -package schemas -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package schemas -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package schemas -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package schemas -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func strPtr(s string) *string {
	return &s
}

func TestGrpcHandler_ListSchemas(t *testing.T) {
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
	schemaData := &IdentitySchemaData{
		IdentitySchemas: []kClient.IdentitySchemaContainer{
			{
				Id:                   strPtr("id1"),
				Schema:               sc,
				AdditionalProperties: ap,
			},
		},
		Tokens: types.NavigationTokens{
			Next: "next",
			Prev: "prev",
		},
		Error: nil,
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name      string
		req       *v0Schemas.ListSchemasReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.ListSchemasResp
		err       error
	}{
		{
			name: "Successful list schemas",
			req: &v0Schemas.ListSchemasReq{
				Size:      nil,
				PageToken: nil,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListSchemas(gomock.Any(), gomock.Any(), gomock.Any()).Return(schemaData, nil)
			},
			want: &v0Schemas.ListSchemasResp{
				Data: []*v0Schemas.Schema{{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				}},
				Status:  http.StatusOK,
				Message: strPtr("List of schemas"),
				XMeta: &v0Types.Pagination{
					Size: 1,
					Next: strPtr("next"),
					Prev: strPtr("prev"),
				},
			},
			err: nil,
		},
		{
			name: "Service list schemas error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListSchemas(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to list schemas, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to list schemas, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.ListSchemas").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.ListSchemas(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("ListSchemas() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ListSchemas() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_GetSchema(t *testing.T) {
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
	schemaData := &IdentitySchemaData{
		IdentitySchemas: []kClient.IdentitySchemaContainer{
			{
				Id:                   strPtr("id1"),
				Schema:               sc,
				AdditionalProperties: ap,
			},
		},
		Tokens: types.NavigationTokens{
			Next: "next",
			Prev: "prev",
		},
		Error: nil,
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name      string
		req       *v0Schemas.GetSchemaReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.GetSchemaResp
		err       error
	}{
		{
			name: "Successful get schema",
			req: &v0Schemas.GetSchemaReq{
				Id: "id1",
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetSchema(gomock.Any(), gomock.Any()).Return(schemaData, nil)
			},
			want: &v0Schemas.GetSchemaResp{
				Data: []*v0Schemas.Schema{{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				}},
				Status:  http.StatusOK,
				Message: strPtr("Schema detail"),
			},
			err: nil,
		},
		{
			name: "Schema not found",
			req: &v0Schemas.GetSchemaReq{
				Id: "id1",
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetSchema(gomock.Any(), gomock.Any()).Return(&IdentitySchemaData{
					IdentitySchemas: []kClient.IdentitySchemaContainer{},
				}, nil)
			},
			want: nil,
			err:  status.Errorf(codes.NotFound, "schema %s not found", "id1"),
		},
		{
			name: "Service get schema error",
			req: &v0Schemas.GetSchemaReq{
				Id: "id1",
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetSchema(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to get schema %s, %v", gomock.Any(), gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to get schema %v, some error", "id1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.GetSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.GetSchema(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("GetSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_CreateSchema(t *testing.T) {
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
	schemaData := &IdentitySchemaData{
		IdentitySchemas: []kClient.IdentitySchemaContainer{
			{
				Id:                   strPtr("id1"),
				Schema:               sc,
				AdditionalProperties: ap,
			},
		},
		Error: nil,
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name      string
		req       *v0Schemas.CreateSchemaReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.CreateSchemaResp
		err       error
	}{
		{
			name: "Successful create schema",
			req: &v0Schemas.CreateSchemaReq{
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().CreateSchema(gomock.Any(), gomock.Any()).Return(schemaData, nil)
			},
			want: &v0Schemas.CreateSchemaResp{
				Data: []*v0Schemas.Schema{{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				}},
				Status:  http.StatusCreated,
				Message: strPtr("Created schema"),
			},
		},
		{
			name: "Service create schema error",
			req: &v0Schemas.CreateSchemaReq{
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().CreateSchema(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to create schema, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to create schema, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.CreateSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.CreateSchema(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("CreateSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("CreateSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_UpdateSchema(t *testing.T) {
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
	schemaData := &IdentitySchemaData{
		IdentitySchemas: []kClient.IdentitySchemaContainer{
			{
				Id:                   strPtr("id1"),
				Schema:               sc,
				AdditionalProperties: ap,
			},
		},
		Error: nil,
	}
	schema, _ := structpb.NewStruct(sc)
	additionalProperties, _ := structpb.NewStruct(ap)

	tests := []struct {
		name      string
		req       *v0Schemas.UpdateSchemaReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.UpdateSchemaResp
		err       error
	}{
		{
			name: "Successful update schema",
			req: &v0Schemas.UpdateSchemaReq{
				Id: "id1",
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return(schemaData, nil)
			},
			want: &v0Schemas.UpdateSchemaResp{
				Data: []*v0Schemas.Schema{{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				}},
				Status:  http.StatusOK,
				Message: strPtr("Updated schema"),
			},
			err: nil,
		},
		{
			name: "Schema ID is empty",
			req: &v0Schemas.UpdateSchemaReq{
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("schema ID is empty")
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "schema ID is empty"),
		},
		{
			name: "Empty request body",
			req: &v0Schemas.UpdateSchemaReq{
				Id:     "id1",
				Schema: nil,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("empty request body")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "empty request body"),
		},
		{
			name: "Mismatch schema ID",
			req: &v0Schemas.UpdateSchemaReq{
				Id: "another_schema_id",
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Errorf("schema ID mismatch, %v != %v", gomock.Any(), gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "schema ID mismatch, %v != %v", "another_schema_id", "id1"),
		},
		{
			name: "Schema does not exist",
			req: &v0Schemas.UpdateSchemaReq{
				Id: "id1",
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&IdentitySchemaData{
						IdentitySchemas: []kClient.IdentitySchemaContainer{},
						Error:           nil,
					},
					nil,
				)
			},
			want: nil,
			err:  status.Errorf(codes.NotFound, "schema %v not found", "id1"),
		},
		{
			name: "Service update schema error",
			req: &v0Schemas.UpdateSchemaReq{
				Id: "id1",
				Schema: &v0Schemas.Schema{
					Id:                   "id1",
					Schema:               schema,
					AdditionalProperties: additionalProperties,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditSchema(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to update schema, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to update schema, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.UpdateSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.UpdateSchema(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("UpdateSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("UpdateSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_RemoveSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       *v0Schemas.RemoveSchemaReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.RemoveSchemaResp
		err       error
	}{
		{
			name: "Successful remove schema",
			req:  &v0Schemas.RemoveSchemaReq{Id: "id1"},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().DeleteSchema(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: &v0Schemas.RemoveSchemaResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed schema"),
			},
			err: nil,
		},
		{
			name: "Schema ID is empty",
			req:  &v0Schemas.RemoveSchemaReq{Id: ""},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("schema ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "schema ID is empty"),
		},
		{
			name: "Service delete schema error",
			req:  &v0Schemas.RemoveSchemaReq{Id: "id1"},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().DeleteSchema(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to remove schema, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to remove schema, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.RemoveSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.RemoveSchema(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("RemoveSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("RemoveSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_GetDefaultSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.GetDefaultSchemaResp
		err       error
	}{
		{
			name: "Successful get default schema",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetDefaultSchema(gomock.Any()).Return(&DefaultSchema{ID: "default"}, nil)
			},
			want: &v0Schemas.GetDefaultSchemaResp{
				Data: &v0Schemas.DefaultSchema{
					Id: "default",
				},
				Status:  http.StatusOK,
				Message: strPtr("Get default schema"),
			},
			err: nil,
		},
		{
			name: "Service get default schema error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetDefaultSchema(gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to get default schema, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to get default schema, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.GetDefaultSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.GetDefaultSchema(context.TODO(), &emptypb.Empty{})

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("GetDefaultSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetDefaultSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_UpdateDefaultSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       *v0Schemas.UpdateDefaultSchemaReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Schemas.UpdateDefaultSchemaResp
		err       error
	}{
		{
			name: "Successful update default schema",
			req: &v0Schemas.UpdateDefaultSchemaReq{
				Schema: &v0Schemas.DefaultSchema{
					Id: "default",
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().UpdateDefaultSchema(gomock.Any(), gomock.Any()).Return(
					&DefaultSchema{
						ID: "default",
					},
					nil,
				)
			},
			want: &v0Schemas.UpdateDefaultSchemaResp{
				Data:    &v0Schemas.DefaultSchema{Id: "default"},
				Status:  http.StatusOK,
				Message: strPtr("Updated default schema"),
			},
			err: nil,
		},
		{
			name: "Service update default schema error",
			req: &v0Schemas.UpdateDefaultSchemaReq{
				Schema: &v0Schemas.DefaultSchema{
					Id: "default",
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().UpdateDefaultSchema(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to update default schema, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to update default schema, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "schemas.GrpcHandler.UpdateDefaultSchema").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.UpdateDefaultSchema(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("UpdateDefaultSchema() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("UpdateDefaultSchema() got = %v, want %v", got, test.want)
			}
		})
	}
}
