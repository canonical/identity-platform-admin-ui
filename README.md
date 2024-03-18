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
- `OPENFGA_API_SCHEME`: scheme for the OpenFGA host variable, either `http` or `https`
- `OPENFGA_API_HOST`: host of the OpenFGA server
- `OPENFGA_API_TOKEN`: token used to interact with OpenFGA server, dictated by OpenFGA server
- `OPENFGA_STORE_ID`: ID of the OpenFGA store the application will talk to 
- `OPENFGA_AUTHORIZATION_MODEL_ID`: ID of the OpenFGA authorization model the application will talk to
- `AUTHORIZATION_ENABLED`: flag defining if the OpenFGA authorization middleware is enabled and, for the time being, if any of the RBAC API are using OpenFGA (to be fixed by https://github.com/canonical/identity-platform-admin-ui/issues/221), default to `false`

## Development setup

As a requirement, please make sure to:
* have `rockcraft`, `yq`, and `make` installed

      snap install rockcraft
      snap install yq
      apt install make

* microk8s is installed with the `registry` addon operating at `localhost:32000` and kubectl configured to use it

      snap install microk8s --classic
      microk8s status --wait-ready
      microk8s enable registry
      # ensure kubectl is configured to use microk8s
      microk8s.kubectl config view --raw > $HOME/.kube/config
      # Alias kubectl so that it can be used by Skaffold
      snap alias microk8s.kubectl kubectl


* ensure [`skaffold`](https://github.com/GoogleContainerTools/skaffold), [`container-structure-test`](https://github.com/GoogleContainerTools/container-structure-test), [`docker`](https://docs.docker.com/engine/install/ubuntu/) and [helm](https://helm.sh/docs/intro/install) are installed according to their documentation

* initialise LXD so that it has a default profile when used by `rockcraft` during `skaffold build`

      sudo lxd init --auto

Run `make dev` to get a working environment in k8s.

To stop any running containers and wipe the container state, run `skaffold delete` from the top of the repository. 

### OpenFGA initialization

The Admin Service comes up with authorization disabled (see `AUTHORIZATION_ENABLED` env var), The env vars `OPENFGA_AUTHORIZATION_MODEL_ID` and `OPENFGA_STORE_ID` which are needed for the correct functioning of the RBAC APIs get set by the job `admin-ui-openfga-setup`, after this has completed a developer is supposed to bounce the deployment to get the application to source the new env vars, setting `AUTHORIZATION_ENABLED` will make sure those endpoints use OpenFGA as a backend instead of a `NoOp implementation` (behaviour will change,  see https://github.com/canonical/identity-platform-admin-ui/issues/221)

```
# Wait for the openfga setup job to complete
kubectl wait --for=condition=complete job/admin-ui-openfga-setup

# Edit the configmap to enable authorization by setting AUTHORIZATION_ENABLED=true
kubectl edit configmap identity-platform-admin-ui

# Restart the admin UI apply the changes
kubectl rollout restart deployment identity-platform-admin-ui
```


K8s jobs don't get deleted on their own so if you wish to make changes to the openfga model, you need to make sure that the job for setting up openfga is deleted before redeploying the the admin UI:

```
kubectl delete job admin-ui-openfga-setup
make dev
```

ensure environment variables in the `identity-platform-admin-ui` configmap reflect the status you want


## Endpoint examples

```shell
> http :8000/api/v0/identities
HTTP/1.1 200 OK
Content-Length: 86
Content-Type: application/json
Date: Wed, 11 Oct 2023 10:05:37 GMT
Vary: Origin

{
    "_meta": {
        "page": 1,
        "size": 100
    },
    "data": [],
    "message": "List of identities",
    "status": 200
}
```

```shell
> http :8000/api/v0/idps      
HTTP/1.1 200 OK
Content-Length: 1520
Content-Type: application/json
Date: Wed, 11 Oct 2023 10:05:43 GMT
Vary: Origin

{
    "_meta": null,
    "data": [
        {
            "apple_private_key": "",
            "apple_private_key_id": "",
            "apple_team_id": "",
            "auth_url": "",
            "client_id": "af675f35-3bd7-4515-88e2-b8032e315f6f",
            "client_secret": "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
            "id": "microsoft_af675f353bd7451588e2b8032e315f6f",
            "issuer_url": "",
            "label": "",
            "mapper_url": "file:///etc/config/kratos/microsoft_schema.jsonnet",
            "microsoft_tenant": "e1574293-28de-4e94-87d5-b61c76fc14e1",
            "provider": "microsoft",
            "requested_claims": null,
            "scope": [
                "profile",
                "email",
                "address",
                "phone"
            ],
            "subject_source": "",
            "token_url": ""
        },
        {
            "apple_private_key": "",
            "apple_private_key_id": "",
            "apple_team_id": "",
            "auth_url": "",
            "client_id": "18fa2999-e6c9-475a-a495-15d933d8e8ce",
            "client_secret": "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
            "id": "google_18fa2999e6c9475aa49515d933d8e8ce",
            "issuer_url": "",
            "label": "",
            "mapper_url": "file:///etc/config/kratos/google_schema.jsonnet",
            "microsoft_tenant": "",
            "provider": "google",
            "requested_claims": null,
            "scope": [
                "profile",
                "email",
                "address",
                "phone"
            ],
            "subject_source": "",
            "token_url": ""
        },
        {
            "apple_private_key": "",
            "apple_private_key_id": "",
            "apple_team_id": "",
            "auth_url": "",
            "client_id": "18fa2999-e6c9-475a-a495-89d941d8e1zy",
            "client_secret": "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
            "id": "aws_18fa2999e6c9475aa49589d941d8e1zy",
            "issuer_url": "",
            "label": "",
            "mapper_url": "file:///etc/config/kratos/google_schema.jsonnet",
            "microsoft_tenant": "",
            "provider": "aws",
            "requested_claims": null,
            "scope": [
                "profile",
                "email",
                "address",
                "phone"
            ],
            "subject_source": "",
            "token_url": ""
        }
    ],
    "message": "List of IDPs",
    "status": 200
}
```

```shell
> http :8000/api/v0/clients     
HTTP/1.1 200 OK
Content-Length: 316
Content-Type: application/json
Date: Wed, 11 Oct 2023 10:05:47 GMT
Vary: Origin

{
    "_links": {
        "first": "/api/v0/clients?page=eyJvZmZzZXQiOiIwIiwidiI6Mn0&size=200",
        "next": "/api/v0/clients?page=eyJvZmZzZXQiOiIyMDAiLCJ2IjoyfQ&size=200",
        "prev": "/api/v0/clients?page=eyJvZmZzZXQiOiItMjAwIiwidiI6Mn0&size=200"
    },
    "_meta": {
        "total_count": "0"
    },
    "data": [],
    "message": "List of clients",
    "status": 200
}
```
