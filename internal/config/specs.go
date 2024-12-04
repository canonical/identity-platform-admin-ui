// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package config

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	OtelGRPCEndpoint string `envconfig:"otel_grpc_endpoint"`
	OtelHTTPEndpoint string `envconfig:"otel_http_endpoint"`
	TracingEnabled   bool   `envconfig:"tracing_enabled" default:"true"`

	LogLevel string `envconfig:"log_level" default:"error"`

	Port        int    `envconfig:"port" default:"8080"`
	ContextPath string `envconfig:"context_path" default:"/"`

	Debug bool `envconfig:"debug" default:"false"`

	KubeconfigFile string `envconfig:"kubeconfig_file"`

	KratosPublicURL string `envconfig:"kratos_public_url" required:"true"`
	KratosAdminURL  string `envconfig:"kratos_admin_url" required:"true"`
	// with no slash suffix
	HydraAdminURL       string `envconfig:"hydra_admin_url" required:"true"`
	OathkeeperPublicURL string `envconfig:"oathkeeper_public_url" required:"true"`

	AuthenticationEnabled       bool     `envconfig:"authentication_enabled" default:"false" validate:"required"`
	OIDCIssuer                  string   `envconfig:"oidc_issuer" validate:"required"`
	OAuth2ClientId              string   `envconfig:"oauth2_client_id" validate:"required"`
	OAuth2ClientSecret          string   `envconfig:"oauth2_client_secret" validate:"required"`
	OAuth2RedirectURI           string   `envconfig:"oauth2_redirect_uri" validate:"required"`
	OAuth2CodeGrantScopes       []string `envconfig:"oauth2_codegrant_scopes" default:"openid,offline_access,profile,email" validate:"required"`
	OAuth2AuthCookiesTTLSeconds int      `envconfig:"oauth2_auth_cookies_ttl_seconds" default:"300" validate:"required"`
	OAuth2UserSessionTTLSeconds int      `envconfig:"oauth2_user_session_ttl_seconds" default:"21600" validate:"required"`

	OAuth2AuthCookiesEncryptionKey  string `envconfig:"oauth2_auth_cookies_encryption_key" required:"true" validate:"required,min=32,max=32"`
	AccessTokenVerificationStrategy string `envconfig:"access_token_verification_strategy" default:"jwks" validate:"oneof=jwks userinfo"`

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

	AuthorizationEnabled     bool `envconfig:"authorization_enabled" default:"false"`
	PayloadValidationEnabled bool `envconfig:"payload_validation_enabled" default:"true"`

	OpenFGAWorkersTotal int `envconfig:"openfga_workers_total" default:"150"`

	MailHost               string `envconfig:"MAIL_HOST" required:"true"`
	MailPort               int    `envconfig:"MAIL_PORT" required:"true"`
	MailUsername           string `envconfig:"MAIL_USERNAME"`
	MailPassword           string `envconfig:"MAIL_PASSWORD"`
	MailFromAddress        string `envconfig:"MAIL_FROM_ADDRESS" required:"true"`
	MailSendTimeoutSeconds int    `envconfig:"MAIL_SEND_TIMEOUT_SECONDS" default:"15"`
}
