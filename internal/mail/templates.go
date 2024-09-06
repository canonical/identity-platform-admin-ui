// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package mail

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
)

var (
	//go:embed html/user-invite.html
	UserCreationInvite embed.FS
)

var (
	templates = map[embed.FS]string{
		UserCreationInvite: "html/user-invite.html",
	}
)

type UserCreationInviteArgs struct {
	InviteUrl    string
	RecoveryCode string
	Email        string
}

func LoadTemplate(templateFS embed.FS) (*template.Template, error) {
	templatePattern, ok := templates[templateFS]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}

	templateName := strings.SplitN(templatePattern, "/", 2)[1]
	return template.New(templateName).ParseFS(templateFS, templatePattern)
}
