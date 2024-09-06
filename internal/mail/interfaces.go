// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package mail

import (
	"context"
	"html/template"

	mail2 "github.com/wneessen/go-mail"
)

type EmailServiceInterface interface {
	Send(context.Context, string, string, *template.Template, any) error
}

type MailClientInterface interface {
	DialAndSendWithContext(context.Context, ...*mail2.Msg) error
}
