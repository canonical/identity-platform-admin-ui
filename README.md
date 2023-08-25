# Identity Platform Admin UI

[![codecov](https://codecov.io/gh/canonical/identity-platform-admin-ui/branch/main/graph/badge.svg?token=Aloh6MWghg)](https://codecov.io/gh/canonical/identity-platform-admin-ui)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/canonical/identity-platform-admin-ui/badge)](https://securityscorecards.dev/viewer/?platform=github.com&org=canonical&repo=identity-platform-admin-ui)
![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/canonical/identity-platform-admin-ui)
[![CI](https://github.com/canonical/identity-platform-admin-ui/actions/workflows/ci.yaml/badge.svg)](https://github.com/canonical/identity-platform-admin-ui/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/canonical/identity-platform-admin-ui.svg)](https://pkg.go.dev/github.com/canonical/identity-platform-admin-ui)

This is the Admin UI for the Canonical Identity Platform.



## Environment Variables

- `OTEL_GRPC_ENDPOINT`: address of the open telemetry grpc endpoint, used for tracing
- `OTEL_HTTP_ENDPOINT`: address of the open telemetry http endpoint, used for tracing (grpc endpoint takes precedence)
- `TRACING_ENABLED`: flag enabling tracing 
- `LOG_LEVEL`: log level, one of `info`,`warn`,`error`,`debug`, defaults to `error`
- `LOG_FILE`: file where to dump logs, defaults to `log.txt`
- `PORT `: http server port, defaults to `8080`
- `DEBUG`: debugging flag for hydra and kratos clients
- `KRATOS_PUBLIC_URL`: Kratos public endpoints address
- `KRATOS_ADMIN_URL`: Kratos admin endpoints address
- `HYDRA_ADMIN_URL`: Hydra admin endpoints address
- `IDP_CONFIGMAP_NAME`: name of the k8s config map containing Identity Providers
- `IDP_CONFIGMAP_NAMESPACE`: namespace of the k8s config map containing Identity Providers
- `SCHEMAS_CONFIGMAP_NAME`: name of the k8s config map containing Identity Schemas
- `SCHEMAS_CONFIGMAP_NAMESPACE`: namespace of the k8s config map containing Identity Schemas

