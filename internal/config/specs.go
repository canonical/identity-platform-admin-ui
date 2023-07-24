package config

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	OtelGRPCEndpoint string `envconfig:"otel_grpc_endpoint"`
	OtelHTTPEndpoint string `envconfig:"otel_http_endpoint"`
	TracingEnabled   bool   `envconfig:"tracing_enabled" default:"true"`

	Debug     bool   `envconfig:"debug" default:"false"`
	KratosURL string `envconfig:"kratos_url" required:"true"`

	LogLevel string `envconfig:"log_level" default:"error"`
	LogFile  string `envconfig:"log_file" default:"log.txt"`

	Port int `envconfig:"port" default:"8080"`

	HydraAdminURL string `envconfig:"hydra_admin_url" required:"true"`
}
