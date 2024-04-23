// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package config

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	OtelGRPCEndpoint string `envconfig:"otel_grpc_endpoint"`
	OtelHTTPEndpoint string `envconfig:"otel_http_endpoint"`
	TracingEnabled   bool   `envconfig:"tracing_enabled" default:"true"`

	LogLevel string `envconfig:"log_level" default:"error"`
	LogFile  string `envconfig:"log_file" default:"log.txt"`

	Port int `envconfig:"port" default:"8080"`

	Debug bool `envconfig:"debug" default:"false"`

	KubeconfigFile      string `envconfig:"kubeconfig_file"`
	KratosPublicURL     string `envconfig:"kratos_public_url" required:"true"`
	KratosAdminURL      string `envconfig:"kratos_admin_url" required:"true"`
	HydraAdminURL       string `envconfig:"hydra_admin_url" required:"true"`
	OathkeeperPublicURL string `envconfig:"oathkeeper_public_url" required:"true"`

	IDPConfigMapName      string `envconfig:"idp_configmap_name" required:"true"`
	IDPConfigMapNamespace string `envconfig:"idp_configmap_namespace" required:"true"`

	SchemasConfigMapName      string `envconfig:"schemas_configmap_name" required:"true"`
	SchemasConfigMapNamespace string `envconfig:"schemas_configmap_namespace" required:"true"`

	RulesConfigMapName      string `envconfig:"rules_configmap_name" required:"true"`
	RulesConfigFileName     string `envconfig:"rules_configmap_file_name" default:"admin_ui_rules.json"`
	RulesConfigMapNamespace string `envconfig:"rules_configmap_namespace" required:"true"`

	ApiScheme string `envconfig:"openfga_api_scheme" default:""`
	ApiHost   string `envconfig:"openfga_api_host"`
	ApiToken  string `envconfig:"openfga_api_token"`
	StoreId   string `envconfig:"openfga_store_id"`
	ModelId   string `envconfig:"openfga_authorization_model_id" default:""`

	AuthorizationEnabled bool `envconfig:"authorization_enabled" default:"false"`

	OpenFGAWorkersTotal int `envconfig:"openfga_workers_total" default:"150"`
}
