// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package mail

import (
	"context"
	"errors"
	"html/template"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package mail -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package mail -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package mail -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package mail -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestEmailService_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempl, _ := LoadTemplate(UserCreationInvite)

	mockArgs := UserCreationInviteArgs{
		InviteUrl:    "test-url",
		RecoveryCode: "test-code",
		Email:        "test-mail",
	}

	tests := []struct {
		name         string
		from         string
		to           string
		template     *template.Template
		templateArgs any
		setupMocks   func(*MockMailClientInterface)
		errMsg       string
	}{
		{
			name:     "Success",
			from:     "example@mail.com",
			to:       "example@mail.com",
			template: mockTempl,
			setupMocks: func(c *MockMailClientInterface) {
				c.EXPECT().DialAndSendWithContext(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:         "FromError",
			from:         "invalid from address",
			to:           "",
			template:     nil,
			templateArgs: nil,
			errMsg:       "failed to parse mail address \"invalid from address\": mail: no angle-addr",
			setupMocks:   func(c *MockMailClientInterface) {},
		},
		{
			name:         "TemplateBodyError",
			from:         "from@example.com",
			to:           "",
			template:     nil,
			templateArgs: nil,
			errMsg:       "template pointer is nil",
			setupMocks:   func(c *MockMailClientInterface) {},
		},
		{
			name:         "ToError",
			from:         "from@example.com",
			to:           "invalid to address",
			template:     mockTempl,
			templateArgs: mockArgs,
			errMsg:       "failed to parse mail address \"invalid to address\": mail: no angle-addr",
			setupMocks:   func(c *MockMailClientInterface) {},
		},
		{
			name:         "SendError",
			from:         "from@example.com",
			to:           "to@example.com",
			template:     mockTempl,
			templateArgs: mockArgs,
			errMsg:       "test-error",
			setupMocks: func(c *MockMailClientInterface) {
				c.EXPECT().DialAndSendWithContext(gomock.Any(), gomock.Any()).Return(errors.New("test-error"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			mockTracer := NewMockTracer(ctrl)
			mockCtx := context.TODO()
			mockTracer.EXPECT().Start(gomock.Any(), "mail.EmailService.Send").Return(mockCtx, trace.SpanFromContext(mockCtx)).AnyTimes()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)

			mockClient := NewMockMailClientInterface(ctrl)

			e := &EmailService{
				from:    tt.from,
				client:  mockClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}

			tt.setupMocks(mockClient)

			if err := e.Send(context.TODO(), tt.to, "test-subject", tt.template, tt.templateArgs); (err != nil) != (tt.errMsg != "") {
				t.Errorf("Send() error, got = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
