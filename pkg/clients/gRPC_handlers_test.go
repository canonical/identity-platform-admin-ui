// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"context"
	"errors"
	v0Clients "github.com/canonical/identity-platform-api/v0/clients"
	v0Types "github.com/canonical/identity-platform-api/v0/http"
	hClient "github.com/ory/hydra-client-go/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reflect"
	"testing"
)

//go:generate mockgen -build_flags=--mod=mod -package clients -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package clients -destination ./mock_clients.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package clients -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package clients -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func strPtr(s string) *string {
	return &s
}

func TestGrpcHandler_ListClients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oauth2Client := hClient.OAuth2Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	client := &v0Clients.Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	tests := []struct {
		name      string
		req       *v0Clients.ListClientsReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Clients.ListClientsResp
		err       error
	}{
		{
			name: "Successful list schemas",
			req: &v0Clients.ListClientsReq{
				ClientName: "",
				Owner:      "",
				Size:       nil,
				PageToken:  "",
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				resp := &ServiceResponse{
					Resp: []OAuth2Client{oauth2Client},
				}
				mockService.EXPECT().ListClients(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: &v0Clients.ListClientsResp{
				Data:    []*v0Clients.Client{client},
				Status:  http.StatusOK,
				Message: strPtr("List of clients"),
				XMeta: &v0Types.Pagination{
					Size:      1,
					PageToken: nil,
					Next:      strPtr(""),
					Prev:      strPtr(""),
				},
			},
			err: nil,
		},
		{
			name: "Service list clients error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().ListClients(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("unexpected internal error, %v", gomock.Any())
			},
			want: nil,
			err:  status.Error(codes.Internal, "unexpected internal error"),
		},
		{
			name: "Hydra list clients error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := &ServiceResponse{
					ServiceError: &ErrorOAuth2{
						Error:            "some error",
						ErrorDescription: "",
						StatusCode:       http.StatusForbidden,
					},
				}
				mockService.EXPECT().ListClients(gomock.Any(), gomock.Any()).Return(res, nil)
				mockLogger.EXPECT().Errorf("failed to list clients, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.PermissionDenied, "failed to list clients"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "clients.GrpcHandler.ListClients").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.ListClients(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("ListClients() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ListClients() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_GetClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oauth2Client := &hClient.OAuth2Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	client := &v0Clients.Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	tests := []struct {
		name      string
		req       *v0Clients.GetClientReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Clients.GetClientResp
		err       error
	}{
		{
			name: "Successfully get client",
			req: &v0Clients.GetClientReq{
				Id: "clientId",
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				resp := &ServiceResponse{
					Resp: oauth2Client,
				}
				mockService.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: &v0Clients.GetClientResp{
				Data:    client,
				Status:  http.StatusOK,
				Message: strPtr("Client detail"),
			},
			err: nil,
		},
		{
			name: "Service get client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("unexpected internal error, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "unexpected internal error"),
		},
		{
			name: "Hydra get client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := &ServiceResponse{
					ServiceError: &ErrorOAuth2{
						Error:            "some error",
						ErrorDescription: "",
						StatusCode:       http.StatusForbidden,
					},
				}
				mockService.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(res, nil)
				mockLogger.EXPECT().Errorf("failed to get client, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.PermissionDenied, "failed to get client"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "clients.GrpcHandler.GetClient").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.GetClient(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("GetClient() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetClient() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_CreateClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oauth2Client := &hClient.OAuth2Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	client := &v0Clients.Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	tests := []struct {
		name      string
		req       *v0Clients.CreateClientReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Clients.CreateClientResp
		err       error
	}{
		{
			name: "Successfully create client",
			req: &v0Clients.CreateClientReq{
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				resp := &ServiceResponse{
					Resp: oauth2Client,
				}
				mockService.EXPECT().CreateClient(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: &v0Clients.CreateClientResp{
				Data:    client,
				Status:  http.StatusCreated,
				Message: strPtr("Created client"),
			},
			err: nil,
		},
		{
			name: "Service create client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().CreateClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("unexpected internal error, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "unexpected internal error"),
		},
		{
			name: "Hydra create client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := &ServiceResponse{
					ServiceError: &ErrorOAuth2{
						Error:            "some error",
						ErrorDescription: "",
						StatusCode:       http.StatusForbidden,
					},
				}
				mockService.EXPECT().CreateClient(gomock.Any(), gomock.Any()).Return(res, nil)
				mockLogger.EXPECT().Errorf("failed to create client, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.PermissionDenied, "failed to create client"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "clients.GrpcHandler.CreateClient").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.CreateClient(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("CreateClient() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("CreateClient() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_UpdateClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oauth2Client := &hClient.OAuth2Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	client := &v0Clients.Client{
		ClientId:                strPtr("clientId"),
		ClientName:              strPtr("clientName"),
		ClientSecret:            strPtr("clientSecret"),
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		RedirectUris:            []string{"http://localhost"},
		Scope:                   strPtr("email"),
		TokenEndpointAuthMethod: strPtr("client_secret_basic"),
	}

	tests := []struct {
		name      string
		req       *v0Clients.UpdateClientReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Clients.UpdateClientResp
		err       error
	}{
		{
			name: "Successfully update client",
			req: &v0Clients.UpdateClientReq{
				Id:     "clientId",
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				resp := &ServiceResponse{
					Resp: oauth2Client,
				}
				mockService.EXPECT().UpdateClient(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: &v0Clients.UpdateClientResp{
				Data:    client,
				Status:  http.StatusCreated,
				Message: strPtr("Updated client"),
			},
			err: nil,
		},
		{
			name: "Client ID is empty",
			req: &v0Clients.UpdateClientReq{
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Error("client ID is empty")
			},
			want: nil,
			err:  status.Error(codes.InvalidArgument, "client ID is empty"),
		},
		{
			name: "Mismatch client ID",
			req: &v0Clients.UpdateClientReq{
				Id:     "id",
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockLogger.EXPECT().Errorf("client ID mismatch, %v != %v", "id", client.GetClientId())
			},
			want: nil,
			err:  status.Errorf(codes.InvalidArgument, "client ID mismatch, %v != %v", "id", client.GetClientId()),
		},
		{
			name: "Service update client error",
			req: &v0Clients.UpdateClientReq{
				Id:     "clientId",
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().UpdateClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("unexpected internal error, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "unexpected internal error"),
		},
		{
			name: "Hydra update client error",
			req: &v0Clients.UpdateClientReq{
				Id:     "clientId",
				Client: client,
			},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := &ServiceResponse{
					ServiceError: &ErrorOAuth2{
						Error:            "some error",
						ErrorDescription: "",
						StatusCode:       http.StatusForbidden,
					},
				}
				mockService.EXPECT().UpdateClient(gomock.Any(), gomock.Any()).Return(res, nil)
				mockLogger.EXPECT().Errorf("failed to update client, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.PermissionDenied, "failed to update client"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "clients.GrpcHandler.UpdateClient").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.UpdateClient(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("UpdateClient() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("UpdateClient() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcHandler_DeleteClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       *v0Clients.RemoveClientReq
		mockSetup func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface)
		want      *v0Clients.RemoveClientResp
		err       error
	}{
		{
			name: "Successful remove client",
			req:  &v0Clients.RemoveClientReq{Id: "id1"},
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := NewServiceResponse()
				mockService.EXPECT().DeleteClient(gomock.Any(), gomock.Any()).Return(res, nil)
			},
			want: &v0Clients.RemoveClientResp{
				Status:  http.StatusOK,
				Message: strPtr("Removed client"),
			},
			err: nil,
		},
		{
			name: "Service delete client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				mockService.EXPECT().DeleteClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
				mockLogger.EXPECT().Errorf("unexpected internal error, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.Internal, "unexpected internal error"),
		},
		{
			name: "Hydra delete client error",
			mockSetup: func(mockService *MockServiceInterface, mockLogger *MockLoggerInterface) {
				res := &ServiceResponse{
					ServiceError: &ErrorOAuth2{
						Error:            "some error",
						ErrorDescription: "",
						StatusCode:       http.StatusForbidden,
					},
				}
				mockService.EXPECT().DeleteClient(gomock.Any(), gomock.Any()).Return(res, nil)
				mockLogger.EXPECT().Errorf("failed to remove client, %v", gomock.Any())
			},
			want: nil,
			err:  status.Errorf(codes.PermissionDenied, "failed to remove client"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := NewMockServiceInterface(ctrl)
			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())
			mockTracer := NewMockTracer(ctrl)
			mockTracer.EXPECT().Start(gomock.Any(), "clients.GrpcHandler.RemoveClient").Return(context.TODO(), mockSpan)

			test.mockSetup(mockService, mockLogger)

			g := NewGrpcHandler(mockService, NewGrpcMapper(mockLogger), mockTracer, mockMonitor, mockLogger)

			got, err := g.RemoveClient(context.TODO(), test.req)

			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("RemoveClient() error = %v, want %v", err, test.err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("RemoveClient() got = %v, want %v", got, test.want)
			}
		})
	}
}
