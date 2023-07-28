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

	KratosURL     string `envconfig:"kratos_url" required:"true"`
	HydraAdminURL string `envconfig:"hydra_admin_url" required:"true"`

	ConfigMapName      string `envconfig:"configmap_name" required:"true"`
	ConfigMapNamespace string `envconfig:"configmap_namespace" required:"true"`
}
