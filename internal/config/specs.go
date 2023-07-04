package config

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	JaegerEndpoint string `envconfig:"jaeger_endpoint"`
	TracingEnabled bool   `envconfig:"tracing_enabled" default:"true"`

	LogLevel string `envconfig:"log_level" default:"error"`
	LogFile  string `envconfig:"log_file" default:"log.txt"`

	Port int `envconfig:"port" default:"8080"`
}
