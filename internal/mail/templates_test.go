// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package mail

import (
	"embed"
	"testing"
)

func TestLoadTemplate(t *testing.T) {

	tests := []struct {
		name         string
		templateFS   embed.FS
		templateName string
		errorMsg     string
	}{
		{
			name:         "Template available",
			templateFS:   UserCreationInvite,
			templateName: "user-invite.html",
		},
		{
			name:       "Template not available",
			templateFS: embed.FS{},
			errorMsg:   "template not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			template, err := LoadTemplate(tt.templateFS)

			if tt.errorMsg != "" {
				if err == nil {
					t.Errorf("expected error != nil")
					return
				}

				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			}

			if tt.errorMsg == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				if template.Name() != tt.templateName {
					t.Errorf("expected templateName %s, got %s", tt.templateName, template.Name())
				}
			}

		})
	}
}
