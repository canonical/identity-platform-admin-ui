package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	fga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
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
	createFgaModelCmd.MarkFlagRequired("fga-store-id")
}

func createModel(apiUrl, apiToken, storeId string) {
	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracer()
	monitor := monitoring.NewNoopMonitor("", logger)
	scheme, host, err := parseURL(apiUrl)
	if err != nil {
		panic(err)
	}
	cfg := fga.NewConfig(scheme, host, storeId, apiToken, "", false, tracer, monitor, logger)
	fgaClient := fga.NewClient(cfg)
	authModelReq := client.ClientWriteAuthorizationModelRequest{
		TypeDefinitions: authorization.AuthModel.TypeDefinitions,
		SchemaVersion:   authorization.AuthModel.SchemaVersion,
		Conditions:      authorization.AuthModel.Conditions,
	}
	modelId, err := fgaClient.WriteModel(context.Background(), &authModelReq)
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
