// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

// createFgaModelCmd represents the createFgaModel command
var createFgaModelCmd = &cobra.Command{
	Use:   "create-fga-model",
	Short: "Creates an openfga model",
	Long:  `Creates an openfga model`,
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl, _ := cmd.Flags().GetString("fga-api-url")
		apiToken, _ := cmd.Flags().GetString("fga-api-token")
		storeId, _ := cmd.Flags().GetString("fga-store-id")
		createModel(apiUrl, apiToken, storeId)
	},
}

func init() {
	rootCmd.AddCommand(createFgaModelCmd)

	createFgaModelCmd.Flags().String("fga-api-url", "", "The openfga API URL")
	createFgaModelCmd.Flags().String("fga-api-token", "", "The openfga API token")
	createFgaModelCmd.Flags().String("fga-store-id", "", "The openfga store to create the model in")
	createFgaModelCmd.MarkFlagRequired("fga-api-url")
	createFgaModelCmd.MarkFlagRequired("fga-api-token")
}

func createModel(apiUrl, apiToken, storeId string) {
	ctx := context.Background()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracer()
	monitor := monitoring.NewNoopMonitor("", logger)
	scheme, host, err := parseURL(apiUrl)
	if err != nil {
		panic(err)
	}
	cfg := openfga.NewConfig(scheme, host, storeId, apiToken, "", false, tracer, monitor, logger)

	fgaClient := openfga.NewClient(cfg)

	if storeId == "" {
		storeId, err = fgaClient.CreateStore(ctx, "identity-admin-ui")

		if err != nil {
			panic(err)
		}

		fgaClient.SetStoreID(ctx, storeId)
	}

	modelId, err := fgaClient.WriteModel(
		context.Background(),
		&client.ClientWriteAuthorizationModelRequest{
			TypeDefinitions: authorization.AuthModel.TypeDefinitions,
			SchemaVersion:   authorization.AuthModel.SchemaVersion,
			Conditions:      authorization.AuthModel.Conditions,
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Created model: %s\n", modelId)
}

func parseURL(s string) (string, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}
