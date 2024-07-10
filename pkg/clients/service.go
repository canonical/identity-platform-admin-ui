// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	hClient "github.com/ory/hydra-client-go/v2"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

type ListClientsRequest struct {
	types.Pagination
	Owner      string `json:"owner,omitempty"`
	ClientName string `json:"client_name,omitempty"`
}

type ErrorOAuth2 struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	StatusCode       int    `json:"-"`
}

type ServiceResponse struct {
	ServiceError *ErrorOAuth2
	Resp         interface{}
	Tokens       types.NavigationTokens
	Meta         map[string]string
}

type Service struct {
	hydra HydraClientInterface
	authz AuthorizerInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) GetClient(ctx context.Context, clientID string) (*ServiceResponse, error) {
	ctx, span := s.tracer.Start(ctx, "hydra.OAuth2Api.GetOAuth2Client")
	defer span.End()

	ret := NewServiceResponse()

	c, resp, err := s.hydra.OAuth2Api().
		GetOAuth2Client(ctx, clientID).
		Execute()

	if err != nil {
		se, err := s.parseServiceError(resp)
		if err != nil {
			return nil, err
		}
		ret.ServiceError = se
	}

	ret.Resp = c
	return ret, nil
}

func (s *Service) DeleteClient(ctx context.Context, clientID string) (*ServiceResponse, error) {
	ctx, span := s.tracer.Start(ctx, "hydra.OAuth2Api.DeleteOAuth2Client")
	defer span.End()

	ret := NewServiceResponse()

	resp, err := s.hydra.OAuth2Api().
		DeleteOAuth2Client(ctx, clientID).
		Execute()

	if err != nil {
		se, err := s.parseServiceError(resp)
		if err != nil {
			return nil, err
		}
		ret.ServiceError = se
		return ret, nil
	}
	s.authz.SetDeleteClientEntitlements(ctx, clientID)
	return ret, nil
}

func (s *Service) CreateClient(ctx context.Context, client *hClient.OAuth2Client) (*ServiceResponse, error) {
	ctx, span := s.tracer.Start(ctx, "hydra.OAuth2Api.CreateOAuth2Client")
	defer span.End()

	ret := NewServiceResponse()

	c, resp, err := s.hydra.OAuth2Api().
		CreateOAuth2Client(ctx).
		OAuth2Client(*client).
		Execute()

	ret.Resp = c
	if err != nil {
		se, err := s.parseServiceError(resp)
		if err != nil {
			return nil, err
		}
		ret.ServiceError = se
		return ret, nil
	}
	s.authz.SetCreateClientEntitlements(ctx, *c.ClientId)

	return ret, nil
}

func (s *Service) UpdateClient(ctx context.Context, client *hClient.OAuth2Client) (*ServiceResponse, error) {
	ctx, span := s.tracer.Start(ctx, "hydra.OAuth2Api.SetOAuth2Client")
	defer span.End()

	ret := NewServiceResponse()

	c, resp, err := s.hydra.OAuth2Api().
		SetOAuth2Client(ctx, *client.ClientId).
		OAuth2Client(*client).
		Execute()

	if err != nil {
		se, err := s.parseServiceError(resp)
		if err != nil {
			return nil, err
		}
		ret.ServiceError = se
	}
	ret.Resp = c
	return ret, nil
}

func (s *Service) ListClients(ctx context.Context, cs *ListClientsRequest) (*ServiceResponse, error) {
	ctx, span := s.tracer.Start(ctx, "hydra.OAuth2Api.ListOAuth2Clients")
	defer span.End()

	ret := NewServiceResponse()

	c, resp, err := s.hydra.OAuth2Api().ListOAuth2Clients(ctx).
		ClientName(cs.ClientName).
		Owner(cs.Owner).
		PageSize(int64(cs.Size)).
		PageToken(cs.PageToken).
		Execute()

	if err != nil {
		se, err := s.parseServiceError(resp)
		if err != nil {
			return nil, err
		}
		ret.ServiceError = se
	}
	ret.Resp = c

	if navTokens, err := types.ParseLinkTokens(resp.Header); err != nil {
		s.logger.Warnf("failed parsing link header: %s", err)
	} else {
		ret.Tokens = navTokens
	}

	ret.Meta["total_count"] = resp.Header.Get("X-Total-Count")

	return ret, nil
}
func (s *Service) UnmarshalClient(data []byte) (*hClient.OAuth2Client, error) {
	c := hClient.NewOAuth2Client()
	err := json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) parseServiceError(r *http.Response) (*ErrorOAuth2, error) {
	// The hydra client does not return any errors, we need to parse the response body and create our
	// own objects.
	// Should we use our objects instead of reusing the ones from the sdk?
	se := new(ErrorOAuth2)

	if r == nil {
		s.logger.Debugf("Got no response from hydra service")
		se.Error = "internal_server_error"
		se.ErrorDescription = "Failed to call hydra service"
		se.StatusCode = http.StatusInternalServerError
		return se, nil
	}

	json_data, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Debugf("Failed to read response body: %s", err)
		return se, err
	}
	err = json.Unmarshal(json_data, se)
	if err != nil {
		s.logger.Debugf("Failed to unmarshal JSON: %s", err)
		return se, err
	}
	se.StatusCode = r.StatusCode

	return se, nil
}

func NewService(hydra HydraClientInterface, authz AuthorizerInterface, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.hydra = hydra
	s.authz = authz

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}

func NewServiceResponse() *ServiceResponse {
	sr := new(ServiceResponse)
	sr.Meta = make(map[string]string)
	return sr
}

func NewListClientsRequest(cn, owner, pageToken string, size int) *ListClientsRequest {
	return &ListClientsRequest{
		ClientName: cn,
		Owner:      owner,
		Pagination: types.Pagination{
			PageToken: pageToken,
			Size:      int64(size),
		},
	}
}
