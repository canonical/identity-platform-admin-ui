// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
	"encoding/json"
	"errors"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"reflect"
	"testing"
)

//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_validation.go -source=../../internal/validation/registry.go

func strPtr(s string) *string {
	return &s
}

func TestGrpcHandler_ListIdps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)
	configurations := []*Configuration{
		{
			ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:        "microsoft",
			Label:           "",
			ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
			IssuerURL:       "",
			AuthURL:         "",
			TokenURL:        "",
			Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
			SubjectSource:   "userinfo",
			TeamId:          "",
			PrivateKeyId:    "",
			PrivateKey:      "",
			Scope:           []string{"profile", "email", "address", "phone"},
			Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
			RequestedClaims: rc,
		},
	}
	idps := []*v0Idps.Idp{
		{
			Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:          "microsoft",
			Label:             strPtr(""),
			ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:      strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
			IssuerUrl:         strPtr(""),
			AuthUrl:           strPtr(""),
			TokenUrl:          strPtr(""),
			MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
			SubjectSource:     strPtr("userinfo"),
			AppleTeamId:       strPtr(""),
			ApplePrivateKeyId: strPtr(""),
			ApplePrivateKey:   strPtr(""),
			Scope:             []string{"profile", "email", "address", "phone"},
			MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
			RequestedClaims:   &requestedClaims,
		},
	}

	tests := []struct {
		name      string
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Idps.ListIdpsResp
		err       error
	}{
		{
			name: "Successful list IDPs",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListResources(gomock.Any()).Return(configurations, nil)
			},
			want: &v0Idps.ListIdpsResp{
				Data:    idps,
				Status:  http.StatusOK,
				Message: strPtr("List of IDPs"),
			},
			err: nil,
		},
		{
			name: "Service list resources error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListResources(gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to list IDPs: %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to list IDPs: some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "idps.GrpcHandler.ListIdps").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.ListIdps(context.TODO(), &emptypb.Empty{})

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("ListIdps() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ListIdps() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_GetIdp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)
	configurations := []*Configuration{
		{
			ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:        "microsoft",
			Label:           "",
			ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
			IssuerURL:       "",
			AuthURL:         "",
			TokenURL:        "",
			Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
			SubjectSource:   "userinfo",
			TeamId:          "",
			PrivateKeyId:    "",
			PrivateKey:      "",
			Scope:           []string{"profile", "email", "address", "phone"},
			Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
			RequestedClaims: rc,
		},
	}
	idps := []*v0Idps.Idp{
		{
			Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:          "microsoft",
			Label:             strPtr(""),
			ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:      strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
			IssuerUrl:         strPtr(""),
			AuthUrl:           strPtr(""),
			TokenUrl:          strPtr(""),
			MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
			SubjectSource:     strPtr("userinfo"),
			AppleTeamId:       strPtr(""),
			ApplePrivateKeyId: strPtr(""),
			ApplePrivateKey:   strPtr(""),
			Scope:             []string{"profile", "email", "address", "phone"},
			MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
			RequestedClaims:   &requestedClaims,
		},
	}

	tests := []struct {
		name      string
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Idps.GetIdpResp
		err       error
	}{
		{
			name: "Successful get IDP",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetResource(gomock.Any(), gomock.Any()).Return(configurations, nil)
			},
			want: &v0Idps.GetIdpResp{
				Data:    idps,
				Status:  http.StatusOK,
				Message: strPtr("IDP details"),
			},
			err: nil,
		},
		{
			name: "Service get IDPs error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetResource(gomock.Any(), idps[0].Id).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to get IDP %s, %v", gomock.Any(), gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to get IDP %v, some error", idps[0].Id),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "idps.GrpcHandler.GetIdp").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.GetIdp(context.TODO(), &v0Idps.GetIdpReq{
				Id: configurations[0].ID,
			})

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("GetIdp() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetIdp() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_CreateIdp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)
	invalidRequestedClaims := `{"user_info": {"given_name": {"essential": true}}`
	configurations := []*Configuration{
		{
			ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:        "microsoft",
			Label:           "",
			ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
			IssuerURL:       "",
			AuthURL:         "",
			TokenURL:        "",
			Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
			SubjectSource:   "userinfo",
			TeamId:          "",
			PrivateKeyId:    "",
			PrivateKey:      "",
			Scope:           []string{"profile", "email", "address", "phone"},
			Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
			RequestedClaims: rc,
		},
	}
	idps := []*v0Idps.Idp{
		{
			Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:          "microsoft",
			Label:             strPtr(""),
			ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:      strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
			IssuerUrl:         strPtr(""),
			AuthUrl:           strPtr(""),
			TokenUrl:          strPtr(""),
			MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
			SubjectSource:     strPtr("userinfo"),
			AppleTeamId:       strPtr(""),
			ApplePrivateKeyId: strPtr(""),
			ApplePrivateKey:   strPtr(""),
			Scope:             []string{"profile", "email", "address", "phone"},
			MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
			RequestedClaims:   &requestedClaims,
		},
	}
	createIdpBody := v0Idps.CreateIdpBody{
		Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
		Provider:          "microsoft",
		Label:             nil,
		ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
		ClientSecret:      nil,
		IssuerUrl:         nil,
		AuthUrl:           nil,
		TokenUrl:          nil,
		MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
		SubjectSource:     strPtr("userinfo"),
		AppleTeamId:       nil,
		ApplePrivateKeyId: nil,
		ApplePrivateKey:   nil,
		Scope:             []string{"profile", "email", "address", "phone"},
		MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
		RequestedClaims:   &requestedClaims,
	}

	tests := []struct {
		name      string
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		req       *v0Idps.CreateIdpReq
		want      *v0Idps.CreateIdpResp
		err       error
	}{
		{
			name: "Successful create IDP",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().CreateResource(gomock.Any(), gomock.Any()).Return(configurations, nil)
			},
			req: &v0Idps.CreateIdpReq{Idp: &createIdpBody},
			want: &v0Idps.CreateIdpResp{
				Data:    idps,
				Status:  http.StatusCreated,
				Message: strPtr("Created IDP"),
			},
			err: nil,
		},
		{
			name: "Invalid requested claims",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Errorf("invalid requested claims: %v", gomock.Any())
			},
			req: &v0Idps.CreateIdpReq{
				Idp: &v0Idps.CreateIdpBody{
					Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
					Provider:          "microsoft",
					Label:             nil,
					ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
					ClientSecret:      nil,
					IssuerUrl:         nil,
					AuthUrl:           nil,
					TokenUrl:          nil,
					MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
					SubjectSource:     strPtr("userinfo"),
					AppleTeamId:       nil,
					ApplePrivateKeyId: nil,
					ApplePrivateKey:   nil,
					Scope:             []string{"profile", "email", "address", "phone"},
					MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
					RequestedClaims:   &invalidRequestedClaims,
				},
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "invalid requested claims: %v", invalidRequestedClaims),
		},
		{
			name: "Service create IDP error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().CreateResource(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to create IDP, %v", gomock.Any())
			},
			req:  &v0Idps.CreateIdpReq{Idp: &createIdpBody},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to create IDP, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "idps.GrpcHandler.CreateIdp").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.CreateIdp(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("CreateIdp() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("CreateIdp() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_UpdateIdp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)
	invalidRequestedClaims := `{"user_info": {"given_name": {"essential": true}}`
	idps := []*v0Idps.Idp{
		{
			Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:          "microsoft",
			Label:             strPtr(""),
			ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:      strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
			IssuerUrl:         strPtr(""),
			AuthUrl:           strPtr(""),
			TokenUrl:          strPtr(""),
			MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
			SubjectSource:     strPtr("userinfo"),
			AppleTeamId:       strPtr(""),
			ApplePrivateKeyId: strPtr(""),
			ApplePrivateKey:   strPtr(""),
			Scope:             []string{"profile", "email", "address", "phone"},
			MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
			RequestedClaims:   &requestedClaims,
		},
	}
	body := &v0Idps.UpdateIdpBody{
		Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
		Provider:          "microsoft",
		Label:             nil,
		ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
		ClientSecret:      nil,
		IssuerUrl:         nil,
		AuthUrl:           nil,
		TokenUrl:          nil,
		MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
		SubjectSource:     strPtr("userinfo"),
		AppleTeamId:       nil,
		ApplePrivateKeyId: nil,
		ApplePrivateKey:   nil,
		Scope:             []string{"profile", "email", "address", "phone"},
		MapperUrl:         nil,
		RequestedClaims:   &requestedClaims,
	}

	configurations := []*Configuration{
		{
			ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
			Provider:        "microsoft",
			Label:           "",
			ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
			ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
			IssuerURL:       "",
			AuthURL:         "",
			TokenURL:        "",
			Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
			SubjectSource:   "userinfo",
			TeamId:          "",
			PrivateKeyId:    "",
			PrivateKey:      "",
			Scope:           []string{"profile", "email", "address", "phone"},
			Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
			RequestedClaims: rc,
		},
	}

	tests := []struct {
		name      string
		req       *v0Idps.UpdateIdpReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Idps.UpdateIdpResp
		err       error
	}{
		{
			name: "Successful update IDP",
			req: &v0Idps.UpdateIdpReq{
				Id:  idps[0].Id,
				Idp: body,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(configurations, nil)
			},
			want: &v0Idps.UpdateIdpResp{
				Data:    idps,
				Status:  http.StatusOK,
				Message: strPtr("Updated IDP"),
			},
			err: nil,
		},
		{
			name: "IDP ID is empty",
			req: &v0Idps.UpdateIdpReq{
				Id:  "",
				Idp: body,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("IDP ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "IDP ID is empty"),
		},
		{
			name: "Empty request body",
			req: &v0Idps.UpdateIdpReq{
				Id:  idps[0].Id,
				Idp: nil,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("empty request body")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "empty request body"),
		},
		{
			name: "Mismatch IDP ID",
			req: &v0Idps.UpdateIdpReq{
				Id:  "another_idp_id",
				Idp: body,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Errorf("IDP ID mismatch, %v != %v", "another_idp_id", idps[0].Id)
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "IDP ID mismatch, %v != %v", "another_idp_id", idps[0].Id),
		},
		{
			name: "Invalid request claims",
			req: &v0Idps.UpdateIdpReq{
				Id: idps[0].Id,
				Idp: &v0Idps.UpdateIdpBody{
					Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
					Provider:          "microsoft",
					Label:             nil,
					ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
					ClientSecret:      nil,
					IssuerUrl:         nil,
					AuthUrl:           nil,
					TokenUrl:          nil,
					MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
					SubjectSource:     strPtr("userinfo"),
					AppleTeamId:       nil,
					ApplePrivateKeyId: nil,
					ApplePrivateKey:   nil,
					Scope:             []string{"profile", "email", "address", "phone"},
					MapperUrl:         nil,
					RequestedClaims:   &invalidRequestedClaims,
				},
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Errorf("invalid requested claims: %v", invalidRequestedClaims)
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "invalid requested claims: %v", invalidRequestedClaims),
		},
		{
			name: "IDP does not exist",
			req: &v0Idps.UpdateIdpReq{
				Id:  idps[0].Id,
				Idp: body,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditResource(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*Configuration{}, nil)
			},
			want: nil,
			err:  status.Errorf(codes.NotFound, "IDP %v not found", idps[0].Id),
		},
		{
			name: "Service update IDP error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().EditResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to update IDP, %v", gomock.Any())
			},
			req: &v0Idps.UpdateIdpReq{
				Id:  idps[0].Id,
				Idp: body,
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to update IDP, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "idps.GrpcHandler.UpdateIdp").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.UpdateIdp(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("UpdateIdp() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("UpdateIdp() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_RemoveIdp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	idpId := "microsoft_af675f353bd7451588e2b8032e315f6f"

	tests := []struct {
		name      string
		req       *v0Idps.RemoveIdpReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Idps.RemoveIdpResp
		err       error
	}{
		{
			name: "Successful remove IDP",
			req:  &v0Idps.RemoveIdpReq{Id: idpId},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().DeleteResource(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: &v0Idps.RemoveIdpResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed IDP"),
			},
			err: nil,
		},
		{
			name: "IDP ID is empty",
			req:  &v0Idps.RemoveIdpReq{Id: ""},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Debug("IDP ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "IDP ID is empty"),
		},
		{
			name: "Service delete resource error",
			req:  &v0Idps.RemoveIdpReq{Id: idpId},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().DeleteResource(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
				mockLogger.EXPECT().Errorf("failed to remove IDP, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "failed to remove IDP, some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "idps.GrpcHandler.RemoveIdp").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.RemoveIdp(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("RemoveIdp() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("RemoveIdp() got = %v, want %v", got, test.want)
			}
		})
	}
}
