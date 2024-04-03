// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	k8s "github.com/canonical/identity-platform-admin-ui/internal/k8s"

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
		kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
		k8sConfigMap, _ := cmd.Flags().GetString("store-k8s-configmap-resource")

		createModel(apiUrl, apiToken, storeId, kubeconfig, k8sConfigMap)
	},
}

func init() {
	rootCmd.AddCommand(createFgaModelCmd)

	createFgaModelCmd.Flags().String("fga-api-url", "", "The openfga API URL")
	createFgaModelCmd.Flags().String("fga-api-token", "", "The openfga API token")
	createFgaModelCmd.Flags().String("fga-store-id", "", "The openfga store to create the model in, if empty one will be created")
	createFgaModelCmd.Flags().String("kubeconfig", "", "The path to the kubeconfig file")
	createFgaModelCmd.Flags().String("store-k8s-configmap-resource", "", "K8s configmap to store created data, format with namespace/configmap-name")
	createFgaModelCmd.MarkFlagRequired("fga-api-url")
	createFgaModelCmd.MarkFlagRequired("fga-api-token")
}

func createModel(apiUrl, apiToken, storeId, kubeconfig, k8sConfigMap string) {
	ctx := context.Background()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracer()
	monitor := monitoring.NewNoopMonitor("", logger)

	scheme, host, err := parseURL(apiUrl)
	if err != nil {
		panic(err)
	}

	// skip validation for openfga object
	cfg := openfga.Config{

		ApiScheme:   scheme,
		ApiHost:     host,
		StoreID:     storeId,
		ApiToken:    apiToken,
		AuthModelID: "",
		Debug:       false,
		Tracer:      tracer,
		Monitor:     monitor,
		Logger:      logger,
	}

	fgaClient := openfga.NewClient(&cfg)

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

	if err := storeValuesK8sConfigMap(ctx, kubeconfig, k8sConfigMap, storeId, modelId); err != nil {
		panic(err)
	}
}

func storeValuesK8sConfigMap(ctx context.Context, kubeconfig, configMap, storeID, modelID string) error {
	fmt.Println(kubeconfig, configMap, storeID, modelID)

	if configMap == "" {
		return nil
	}

	k8sClient, err := k8s.NewCoreV1Client(kubeconfig)

	if err != nil {
		return err
	}

	parts := strings.Split(configMap, "/")

	if len(parts) != 2 {
		return fmt.Errorf("invalid format for configmap resource %s: expected namespace/name", configMap)
	}
	cmNamespace, cmName := parts[0], parts[1]

	cm, err := k8sClient.ConfigMaps(cmNamespace).Get(ctx, cmName, metaV1.GetOptions{})

	if err != nil {
		return err
	}

	cm.Data["OPENFGA_STORE_ID"] = storeID
	cm.Data["OPENFGA_AUTHORIZATION_MODEL_ID"] = modelID

	if _, err = k8sClient.ConfigMaps(cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {
		return err
	}

	fmt.Printf("Configmap updated successfully: %s\n", configMap)

	return nil
}

func parseURL(s string) (string, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}
