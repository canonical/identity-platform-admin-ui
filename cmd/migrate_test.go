// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCustomValidArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "Empty args",
			args:      []string{},
			wantError: false,
		},
		{
			name:      "Valid up",
			args:      []string{"up"},
			wantError: false,
		},
		{
			name:      "Valid down without version",
			args:      []string{"down"},
			wantError: false,
		},
		{
			name:      "Valid down with version 0",
			args:      []string{"down", "0"},
			wantError: false,
		},
		{
			name:      "Valid down with version 5",
			args:      []string{"down", "5"},
			wantError: false,
		},
		{
			name:      "Valid status",
			args:      []string{"status"},
			wantError: false,
		},
		{
			name:      "Invalid command",
			args:      []string{"invalid"},
			wantError: true,
		},
		{
			name:      "Invalid second arg with up",
			args:      []string{"up", "extra"},
			wantError: true,
		},
		{
			name:      "Invalid second arg with status",
			args:      []string{"status", "extra"},
			wantError: true,
		},
		{
			name:      "Invalid non-down command with 2 args",
			args:      []string{"up", "1"},
			wantError: true,
		},
		{
			name:      "Down with negative number",
			args:      []string{"down", "-1"},
			wantError: true,
		},
		{
			name:      "Down with non-numeric second arg",
			args:      []string{"down", "abc"},
			wantError: true,
		},
		{
			name:      "Too many args",
			args:      []string{"down", "1", "extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := customValidArgs()(cmd, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateCLIArgs() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
