// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	"github.com/canonical/identity-platform-admin-ui/migrations"
)

// migrateCmd performs DB migrations
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations`,
	Args:  customValidArgs(),
	Run:   runMigrate(),
}

func customValidArgs() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}

		if err := cobra.RangeArgs(0, 2)(cmd, args); err != nil {
			return err
		}

		first := args[0]
		switch first {
		case "up", "down", "status":
			// valid first argument
		default:
			return fmt.Errorf("invalid first argument: %q", first)
		}

		// If two arguments are provided, the first must be "down" and second a non-negative int
		if len(args) == 2 {
			if first != "down" {
				return fmt.Errorf("invalid argument combination: %q", args)
			}

			if version, err := strconv.Atoi(args[1]); err != nil || version < 0 {
				return fmt.Errorf("invalid version number: %q", args[1])
			}
		}

		return nil
	}
}

func runMigrate() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		command := "up"
		if len(args) > 0 {
			command = args[0]
		}

		version := -1
		if len(args) > 1 {
			version, _ = strconv.Atoi(args[1])
		}

		dsn, _ := cmd.Flags().GetString("dsn")

		if err := migrate(cmd.Context(), dsn, command, version); err != nil {
			panic(err)
		}
	}
}

func init() {
	migrateCmd.Flags().String("dsn", "", "PostgreSQL DSN connection string")
	_ = migrateCmd.MarkFlagRequired("dsn")

	rootCmd.AddCommand(migrateCmd)
}

func migrate(ctx context.Context, dsn, command string, version int) error {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("DSN validation failed, shutting down, err: %v", err)
	}

	db := stdlib.OpenDB(*config)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("DB connection failed, shutting down, err: %v", err)
	}
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	switch command {
	case "up":
		// up to most recent migration
		err = goose.UpContext(ctx, db, ".", goose.WithNoColor(true))
	case "down":
		if version == -1 {
			// no version arg was passed, downgrade to previous version
			return goose.DownContext(ctx, db, ".", goose.WithNoColor(true))
		}

		err = goose.DownToContext(ctx, db, ".", int64(version), goose.WithNoColor(true))
	case "status":
		err = goose.StatusContext(ctx, db, ".", goose.WithNoColor(true))
	}

	return err
}
