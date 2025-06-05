// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	"github.com/canonical/identity-platform-admin-ui/migrations"
)

// migrateCmd performs DB migrations
var migrateCmd = &cobra.Command{
	Use:        "migrate",
	SuggestFor: nil,
	Short:      "Run database migrations",
	Long:       `Run database migrations`,
	Example:    "",
	ValidArgs:  []string{"up", "down", "status"},
	Args:       cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := "up"
		if len(args) > 0 {
			command = args[0]
		}

		dsn, _ := cmd.Flags().GetString("dsn")

		if err := migrate(cmd.Context(), dsn, command); err != nil {
			panic(err)
		}
	},
	CompletionOptions: cobra.CompletionOptions{},
}

func init() {
	migrateCmd.Flags().String("dsn", "", "PostgreSQL DSN connection string")
	_ = migrateCmd.MarkFlagRequired("dsn")

	rootCmd.AddCommand(migrateCmd)
}

func migrate(ctx context.Context, dsn, command string) error {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("DSN validation failed, shutting down, dsn: %s, err: %v", dsn, err)
	}

	db := stdlib.OpenDB(*config)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("DB connection failed, shutting down, dsn: %s, err: %v", dsn, err)
	}
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	switch command {
	case "up":
		err = goose.UpContext(ctx, db, ".", goose.WithNoColor(true))
	case "down":
		err = goose.DownToContext(ctx, db, ".", 0, goose.WithNoColor(true))
	case "status":
		err = goose.StatusContext(ctx, db, ".", goose.WithNoColor(true))
	}

	return err
}
