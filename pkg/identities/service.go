// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	kClient "github.com/ory/kratos-client-go"
	"github.com/tomnomnom/linkheader"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

type Service struct {
	kratos kClient.IdentityAPI

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// TODO @shipperizer worth offloading to a different place as it's going to be reused
type PaginationTokens struct {
	First string
	Prev  string
	Next  string
}

type IdentityData struct {
	Identities []kClient.Identity
	Tokens     PaginationTokens
	Error      *kClient.GenericError
}

// TODO @shipperizer verify during integration test if this is actually the format
type KratosError struct {
	Error *kClient.GenericError `json:"error,omitempty"`
}

func (s *Service) buildListRequest(ctx context.Context, size int64, token, credID string) kClient.IdentityAPIListIdentitiesRequest {
	r := s.kratos.ListIdentities(ctx).PageToken(token).PageSize(size)

	if credID != "" {
		r = r.CredentialsIdentifier(credID)
	}

	return r
}

func (s *Service) parseLinkURL(linkURL string) string {
	u, err := url.Parse(linkURL)

	if err != nil {
		s.logger.Errorf("failed to parse link header successfully: %s", err)
		return ""
	}

	return u.Query().Get("page_token")
}

func (s *Service) parsePagination(r *http.Response) PaginationTokens {
	links := linkheader.Parse(r.Header.Get("Link"))

	pagination := PaginationTokens{}

	for _, link := range links {
		switch link.Rel {
		case "first":
			pagination.First = s.parseLinkURL(link.URL)
		case "next":
			pagination.Next = s.parseLinkURL(link.URL)
		case "prev":
			pagination.Prev = s.parseLinkURL(link.URL)
		}
	}

	return pagination
}

func (s *Service) parseError(r *http.Response) *kClient.GenericError {
	gerr := KratosError{Error: kClient.NewGenericErrorWithDefaults()}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(body, &gerr); err != nil {
		gerr.Error.SetMessage("unable to parse kratos error response")
		gerr.Error.SetCode(http.StatusInternalServerError)
	}

	return gerr.Error
}

// TODO @shipperizer fix pagination
// shipperizer in ~/shipperizer/identity-platform-admin-ui on IAM-782 ● ● λ http :18000/admin/identities
// HTTP/1.1 200 OK
// Cache-Control: private, no-cache, no-store, must-revalidate
// Content-Length: 3
// Content-Type: application/json; charset=utf-8
// Date: Thu, 04 Apr 2024 14:08:03 GMT
// Link: <http://kratos-admin.default.svc.cluster.local/identities?page=0&page_size=250&page_token=eyJvZmZzZXQiOiIwIiwidiI6Mn0&per_page=250>; rel="first",<http://kratos-admin.default.svc.cluster.local/identities?page=1&page_size=250&page_token=eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ&per_page=250>; rel="next",<http://kratos-admin.default.svc.cluster.local/identities?page=-1&page_size=250&page_token=eyJvZmZzZXQiOiItMjUwIiwidiI6Mn0&per_page=250>; rel="prev"
// X-Total-Count: 0
// []
func (s *Service) ListIdentities(ctx context.Context, size int64, token, credID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.ListIdentities")
	defer span.End()

	identities, rr, err := s.kratos.ListIdentitiesExecute(
		s.buildListRequest(ctx, size, token, credID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	data.Tokens = s.parsePagination(rr)
	data.Identities = identities

	// TODO @shipperizer check if identities is defaulting to empty slice inside kratos-client
	if data.Identities == nil {
		data.Identities = make([]kClient.Identity, 0)
	}

	return data, err
}

func (s *Service) GetIdentity(ctx context.Context, ID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.GetIdentity")
	defer span.End()

	identity, rr, err := s.kratos.GetIdentityExecute(
		s.kratos.GetIdentity(ctx, ID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	return data, err
}

func (s *Service) CreateIdentity(ctx context.Context, bodyID *kClient.CreateIdentityBody) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.CreateIdentity")
	defer span.End()

	if bodyID == nil {
		err := fmt.Errorf("no identity data passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	identity, rr, err := s.kratos.CreateIdentityExecute(
		s.kratos.CreateIdentity(ctx).CreateIdentityBody(*bodyID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	return data, err
}

func (s *Service) UpdateIdentity(ctx context.Context, ID string, bodyID *kClient.UpdateIdentityBody) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.UpdateIdentity")
	defer span.End()
	if ID == "" {
		err := fmt.Errorf("no identity ID passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	if bodyID == nil {
		err := fmt.Errorf("no identity body passed")

		data := new(IdentityData)
		data.Identities = []kClient.Identity{}
		data.Error = s.parseError(nil)
		data.Error.SetMessage(err.Error())

		s.logger.Error(err)

		return data, err
	}

	identity, rr, err := s.kratos.UpdateIdentityExecute(
		s.kratos.UpdateIdentity(ctx, ID).UpdateIdentityBody(*bodyID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	if identity != nil {
		data.Identities = []kClient.Identity{*identity}
	} else {
		data.Identities = []kClient.Identity{}
	}

	return data, err
}

func (s *Service) DeleteIdentity(ctx context.Context, ID string) (*IdentityData, error) {
	ctx, span := s.tracer.Start(ctx, "kratos.IdentityAPI.DeleteIdentity")
	defer span.End()

	rr, err := s.kratos.DeleteIdentityExecute(
		s.kratos.DeleteIdentity(ctx, ID),
	)

	data := new(IdentityData)

	if err != nil {
		s.logger.Error(err)
		data.Error = s.parseError(rr)
	}

	data.Identities = []kClient.Identity{}

	return data, err
}

func NewService(kratos kClient.IdentityAPI, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	s.kratos = kratos

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
