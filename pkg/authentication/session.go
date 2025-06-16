// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	monitoring "github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	kClient "github.com/ory/kratos-client-go"
)

type SessionManager struct {
	kratos       kClient.IdentityAPI
	kratosPublic kClient.FrontendAPI

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

type SessionData struct {
	Session kClient.Session
	Error   *kClient.GenericError
}

func (s *SessionManager) cookiesToString(cookies []*http.Cookie) string {
	var ret = make([]string, len(cookies))
	for i, c := range cookies {
		ret[i] = fmt.Sprintf("%s=%s", c.Name, c.Value)
	}
	return strings.Join(ret, "; ")
}

func (s *SessionManager) GetIdentitySession(ctx context.Context, cookies []*http.Cookie) (*SessionData, error) {
	ctx, span := s.tracer.Start(ctx, "authentication.SessionManager.GetIdentitySession")
	defer span.End()

	// GET /sessions/whoami
	session, rr, err := s.kratosPublic.
		ToSession(ctx).
		Cookie(s.cookiesToString(cookies)).
		Execute()

	data := new(SessionData)
	if err != nil {
		data.Error = kratos.ParseKratosError(rr)
		s.logger.Errorf("failed to disable kratos session: %v, error message: %s", err, data.Error.Message)
		return data, err
	}

	if session != nil {
		data.Session = *session
	}

	return data, nil
}

func (s *SessionManager) DisableSession(ctx context.Context, sessionID string) (*SessionData, error) {
	ctx, span := s.tracer.Start(ctx, "authentication.SessionManager.DisableSession")
	defer span.End()

	// DEL /admin/sessions/{id}
	rr, err := s.kratos.DisableSession(ctx, sessionID).Execute()

	data := new(SessionData)
	if err != nil {
		data.Error = kratos.ParseKratosError(rr)
		s.logger.Errorf("failed to disable kratos session: %v, error message: %s", err, data.Error.Message)
		return data, err
	}

	return data, err
}

func NewSessionManagerService(kratos kClient.IdentityAPI, kratosPublic kClient.FrontendAPI, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *SessionManager {
	s := new(SessionManager)

	s.kratos = kratos
	s.kratosPublic = kratosPublic

	s.tracer = tracer
	s.monitor = monitor
	s.logger = logger

	return s
}
