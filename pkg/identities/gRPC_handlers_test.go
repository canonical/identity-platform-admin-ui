// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	v0Types "github.com/canonical/identity-platform-api/v0/http"
	v0Identities "github.com/canonical/identity-platform-api/v0/identities"
	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
)

//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func ptr(s string) *string {
	return &s
}

func TestGrpcHandler_ListIdentities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"

	tests := []struct {
		name      string
		req       *v0Identities.ListIdentitiesReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Identities.ListIdentitiesResp
		err       error
	}{
		{
			name: "Success",
			req: &v0Identities.ListIdentitiesReq{
				CredID: "123",
				Size:   ptr("2"),
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListIdentities(gomock.Any(), int64(2), "", "123").
					Return(&IdentityData{
						Identities: []kClient.Identity{
							{Id: "1", State: &mockState, OrganizationId: kClient.NullableString{}},
							{Id: "2", State: &mockState, OrganizationId: kClient.NullableString{}},
						},
						Tokens: types.NavigationTokens{Next: "next", Prev: "prev"},
					}, nil)
			},
			want: &v0Identities.ListIdentitiesResp{
				Data: []*v0Identities.Identity{
					{Id: "1", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
					{Id: "2", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
				},
				Status:  http.StatusOK,
				Message: ptr("List of identities"),
				XMeta: &v0Types.Pagination{
					Size: 2,
					Next: ptr("next"),
					Prev: ptr("prev"),
				},
			},
			err: nil,
		},
		{
			name: "empty credID",
			req:  &v0Identities.ListIdentitiesReq{},
			mockSetup: func(_ *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Debug("credID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "credID is empty"),
		},
		{
			name: "invalid page size format",
			req: &v0Identities.ListIdentitiesReq{
				CredID: "123",
				Size:   ptr("not-an-int"),
			},
			mockSetup: func(_ *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Debug("page size parameter is not an int")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "page size parameter is not an int, strconv.ParseInt: parsing \"not-an-int\": invalid syntax"),
		},
		{
			name: "ListIdentities service returns error",
			req: &v0Identities.ListIdentitiesReq{
				CredID: "123",
				Size:   ptr("10"),
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListIdentities(gomock.Any(), int64(10), "", "123").
					Return(nil, fmt.Errorf("db error"))
				mockLogger.EXPECT().Errorf("failed to list identities, %v", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to list identities, db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})
			mockTracer.EXPECT().Start(gomock.Any(), "identities.GrpcHandler.ListIdentities").Return(mockCtx, mockSpan)

			tt.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.ListIdentities(context.TODO(), tt.req)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("ListIdentities() error = %v, want %v", err, tt.err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListIdentities() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcHandler_GetIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"

	tests := []struct {
		name      string
		req       *v0Identities.GetIdentityReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Identities.GetIdentityResp
		err       error
	}{
		{
			name: "Success",
			req:  &v0Identities.GetIdentityReq{Id: "identity-1"},
			mockSetup: func(mockService *MockServiceInterface, _ *MockLoggerInterface) {
				mockService.EXPECT().
					GetIdentity(gomock.Any(), "identity-1").
					Return(&IdentityData{
						Identities: []kClient.Identity{
							{Id: "identity-1", State: &mockState, OrganizationId: kClient.NullableString{}},
						},
					}, nil)
			},
			want: &v0Identities.GetIdentityResp{
				Data: []*v0Identities.Identity{
					{Id: "identity-1", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
				},
				Status:  http.StatusOK,
				Message: ptr("Identity detail"),
			},
			err: nil,
		},
		{
			name: "Service returns error",
			req:  &v0Identities.GetIdentityReq{Id: "identity-2"},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().
					GetIdentity(gomock.Any(), "identity-2").
					Return(nil, fmt.Errorf("db failure"))

				mockLogger.EXPECT().
					Errorf("failed to get identity %s, %v", "identity-2", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to get identity identity-2, db failure"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})
			mockTracer.EXPECT().Start(gomock.Any(), "identities.GrpcHandler.GetIdentity").Return(mockCtx, mockSpan)

			tt.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.GetIdentity(context.TODO(), tt.req)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("GetIdentity() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcHandler_CreateIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"
	mockSchemaId := "new-schema-id"

	tests := []struct {
		name      string
		req       *v0Identities.CreateIdentityReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Identities.CreateIdentityResp
		err       error
	}{
		{
			name: "Success",
			req: &v0Identities.CreateIdentityReq{
				Identity: &v0Identities.CreateIdentityBody{
					SchemaId: mockSchemaId,
					State:    mockState,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, _ *MockLoggerInterface) {
				mockService.EXPECT().
					CreateIdentity(gomock.Any(), &kClient.CreateIdentityBody{
						SchemaId: mockSchemaId,
						State:    &mockState,
					}).
					Return(&IdentityData{
						Identities: []kClient.Identity{{SchemaId: "new-schema-id", State: &mockState, OrganizationId: kClient.NullableString{}}},
					}, nil)

				mockService.EXPECT().
					SendUserCreationEmail(gomock.Any(), &kClient.Identity{SchemaId: "new-schema-id", State: &mockState, OrganizationId: kClient.NullableString{}}).
					Return(nil)
			},
			want: &v0Identities.CreateIdentityResp{
				Data: []*v0Identities.Identity{
					{SchemaId: "new-schema-id", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
				},
				Status:  http.StatusCreated,
				Message: ptr("Created identity"),
			},
			err: nil,
		},
		{
			name: "CreateIdentity fails",
			req: &v0Identities.CreateIdentityReq{
				Identity: &v0Identities.CreateIdentityBody{
					SchemaId: mockSchemaId,
					State:    mockState,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().
					CreateIdentity(gomock.Any(), &kClient.CreateIdentityBody{
						SchemaId: mockSchemaId,
						State:    &mockState,
					}).
					Return(nil, fmt.Errorf("creation failed"))

				mockLogger.EXPECT().
					Errorf("failed to create identity, %v", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to create identity, creation failed"),
		},
		{
			name: "SendUserCreationEmail fails",
			req: &v0Identities.CreateIdentityReq{
				Identity: &v0Identities.CreateIdentityBody{
					SchemaId: mockSchemaId,
					State:    mockState,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().
					CreateIdentity(gomock.Any(), &kClient.CreateIdentityBody{
						SchemaId: mockSchemaId,
						State:    &mockState,
					}).
					Return(&IdentityData{
						Identities: []kClient.Identity{{Id: "created-id"}},
					}, nil)

				mockService.EXPECT().
					SendUserCreationEmail(gomock.Any(), &kClient.Identity{Id: "created-id"}).
					Return(fmt.Errorf("email failed"))

				mockLogger.EXPECT().
					Errorf("failed to send user creation email, %v", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to send user creation email, email failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})
			mockTracer.EXPECT().Start(gomock.Any(), "identities.GrpcHandler.CreateIdentity").Return(mockCtx, mockSpan)

			tt.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.CreateIdentity(context.TODO(), tt.req)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("CreateIdentity() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcHandler_UpdateIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"
	mockSchemaId := "new-schema-id"

	tests := []struct {
		name      string
		req       *v0Identities.UpdateIdentityReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Identities.UpdateIdentityResp
		err       error
	}{
		{
			name: "Success",
			req: &v0Identities.UpdateIdentityReq{
				Id: "123",
				Identity: &v0Identities.UpdateIdentityBody{
					SchemaId: mockSchemaId,
					State:    mockState,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, _ *MockLoggerInterface) {
				mockService.EXPECT().
					UpdateIdentity(gomock.Any(), "123", &kClient.UpdateIdentityBody{
						SchemaId: mockSchemaId,
						State:    mockState,
					}).
					Return(&IdentityData{
						Identities: []kClient.Identity{{Id: "123", State: &mockState, OrganizationId: kClient.NullableString{}}},
					}, nil)
			},
			want: &v0Identities.UpdateIdentityResp{
				Data: []*v0Identities.Identity{
					{Id: "123", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
				},
				Status:  http.StatusOK,
				Message: ptr("Updated identity"),
			},
			err: nil,
		},
		{
			name: "Empty ID",
			req: &v0Identities.UpdateIdentityReq{
				Id: "",
			},
			mockSetup: func(_ *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().
					Debug("identity ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "identity ID is empty"),
		},
		{
			name: "UpdateIdentity fails",
			req: &v0Identities.UpdateIdentityReq{
				Id: "123",
				Identity: &v0Identities.UpdateIdentityBody{
					SchemaId: mockSchemaId,
					State:    mockState,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().
					UpdateIdentity(gomock.Any(), "123", &kClient.UpdateIdentityBody{
						SchemaId: mockSchemaId,
						State:    mockState,
					}).
					Return(nil, fmt.Errorf("update failed"))

				mockLogger.EXPECT().
					Errorf("failed to update identity %s, %v", "123", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to update identity 123, update failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})
			mockTracer.EXPECT().Start(gomock.Any(), "identities.GrpcHandler.UpdateIdentity").Return(mockCtx, mockSpan)

			tt.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.UpdateIdentity(context.TODO(), tt.req)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("UpdateIdentity() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcHandler_RemoveIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockState := "mock-state"

	tests := []struct {
		name      string
		req       *v0Identities.RemoveIdentityReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Identities.RemoveIdentityResp
		err       error
	}{
		{
			name: "Success",
			req:  &v0Identities.RemoveIdentityReq{Id: "123"},
			mockSetup: func(mockService *MockServiceInterface, _ *MockLoggerInterface) {
				mockService.EXPECT().
					DeleteIdentity(gomock.Any(), "123").
					Return(&IdentityData{
						Identities: []kClient.Identity{{Id: "123", State: &mockState, OrganizationId: kClient.NullableString{}}},
					}, nil)
			},
			want: &v0Identities.RemoveIdentityResp{
				Data: []*v0Identities.Identity{
					{Id: "123", State: &mockState, OrganizationId: &v0Identities.NullableString{}},
				},
				Status:  http.StatusOK,
				Message: ptr("Identity deleted"),
			},
			err: nil,
		},
		{
			name: "Empty ID",
			req:  &v0Identities.RemoveIdentityReq{Id: ""},
			mockSetup: func(_ *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().
					Debug("identity ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "identity ID is empty"),
		},
		{
			name: "DeleteIdentity fails",
			req:  &v0Identities.RemoveIdentityReq{Id: "123"},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().
					DeleteIdentity(gomock.Any(), "123").
					Return(nil, fmt.Errorf("delete error"))

				mockLogger.EXPECT().
					Errorf("failed to delete identity %s, %v", "123", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "failed to delete identity 123, delete error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockCtx := authentication.PrincipalContext(context.TODO(), &authentication.UserPrincipal{Email: "test-user"})
			mockTracer.EXPECT().Start(gomock.Any(), "identities.GrpcHandler.RemoveIdentity").Return(mockCtx, mockSpan)

			tt.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.RemoveIdentity(context.TODO(), tt.req)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("RemoveIdentity() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}
