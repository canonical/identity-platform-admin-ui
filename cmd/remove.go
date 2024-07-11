// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an admin user",
	Long:  `Remove an admin user.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl, _ := cmd.Flags().GetString("fga-api-url")
		apiToken, _ := cmd.Flags().GetString("fga-api-token")
		storeId, _ := cmd.Flags().GetString("fga-store-id")
		modelId, _ := cmd.Flags().GetString("fga-model-id")
		user, _ := cmd.Flags().GetString("user")

		removeAdmin(apiUrl, apiToken, storeId, modelId, user)
	},
}

func init() {
	adminCmd.AddCommand(removeCmd)

	removeCmd.Flags().String("fga-api-url", "", "The openfga API URL")
	removeCmd.Flags().String("fga-api-token", "", "The openfga API token")
	removeCmd.Flags().String("fga-store-id", "", "The openfga store")
	removeCmd.Flags().String("fga-model-id", "", "The openfga model")
	removeCmd.Flags().String("user", "", "The admin user name, if not specified an autogenerated ID will be used")
	removeCmd.MarkFlagRequired("fga-api-url")
	removeCmd.MarkFlagRequired("fga-api-token")
	removeCmd.MarkFlagRequired("fga-store-id")
	removeCmd.MarkFlagRequired("user")
}

func removeAdmin(apiUrl, apiToken, storeId, ModelId, user string) {
	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracer()
	monitor := monitoring.NewNoopMonitor("", logger)
	scheme, host, err := parseURL(apiUrl)
	if err != nil {
		panic(err)
	}
	cfg := openfga.NewConfig(scheme, host, storeId, apiToken, "", false, tracer, monitor, logger)
	fgaClient := openfga.NewClient(cfg)
	wpool := pool.NewWorkerPool(1, tracer, monitor, logger)
	auth := authorization.NewAuthorizer(fgaClient, wpool, tracer, monitor, logger)

	err = auth.RemoveAdmin(context.Background(), user)
	if err != nil {
		fmt.Printf("failed to remove admin: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Removed admin user: %s\n", user)
}
