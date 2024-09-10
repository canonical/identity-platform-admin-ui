// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package mail

import (
	"context"
	"html/template"
	"time"

	"github.com/wneessen/go-mail"
	"go.opentelemetry.io/otel/trace"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
)

type Config struct {
	Host        string `validate:"required"`
	Port        int    `validate:"required"`
	Username    string
	Password    string
	FromAddress string `validate:"required"`
	SendTimeout time.Duration
}

func NewConfig(host string, port int, username, password, from string, sendTimeout int) *Config {
	c := new(Config)

	c.Host = host
	c.Port = port
	c.Username = username
	c.Password = password
	c.FromAddress = from
	c.SendTimeout = time.Duration(sendTimeout) * time.Second

	return c
}

type EmailService struct {
	from   string
	client MailClientInterface

	tracer  trace.Tracer
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (e *EmailService) Send(ctx context.Context, to, subject string, template *template.Template, templateArgs any) error {
	ctx, span := e.tracer.Start(ctx, "mail.EmailService.Send")
	defer span.End()

	msg := mail.NewMsg()

	if err := msg.From(e.from); err != nil {
		return err
	}

	if err := msg.SetBodyHTMLTemplate(template, templateArgs); err != nil {
		return err
	}

	if err := msg.To(to); err != nil {
		return err
	}

	msg.Subject(subject)

	return e.client.DialAndSendWithContext(ctx, msg)
}

func NewEmailService(config *Config, tracer trace.Tracer, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *EmailService {
	s := new(EmailService)
	s.from = config.FromAddress

	var err error
	mailOpts := []mail.Option{
		mail.WithPort(config.Port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithTimeout(config.SendTimeout),
	}

	// treat smtp connection as authenticated only if username is passed
	if config.Username != "" {
		mailOpts = append(
			mailOpts,
			[]mail.Option{mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(config.Username), mail.WithPassword(config.Password)}...,
		)
	}

	s.client, err = mail.NewClient(
		config.Host,
		mailOpts...,
	)

	if err != nil {
		logger.Fatalf("failed to create email client: %s", err)
	}

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
