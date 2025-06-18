# Developing Admin UI

## Prerequisites

- when both Docker and LXD are installed, please make sure to run this as a
  general setup step:

  ```shell
  # this will allow connections from the DOCKER-USER which would otherwise be rejected
  sudo iptables -I DOCKER-USER  -j ACCEPT
  ```

- **don't rely** on `microk8s` binary of `kubectl` (which is invoked
  with `microk8s.kubectl`), but make sure you install `kubectl` following
  [official docs](https://kubernetes.io/docs/tasks/tools/)
- make sure you use the `skopeo` binary from the `rockcraft` installation, you
  can use a symlink

  ```shell
  sudo ln -s /usr/bin/skopeo /snap/rockcraft/current/bin/skopeo
  ```

## Deploying locally

OpenFGA will have a new authorization model and a new store.
This means that, when using environment variables to have Admin UI backend point
to OpenFGA instance, you need to update both

- `OPENFGA_STORE_ID`
- `OPENFGA_AUTHORIZATION_MODEL_ID`

Taking the correct values from the config map.
To do that, you can use `yq` and `kubectl`:

```shell
$ kubectl get cm identity-platform-admin-ui -o yaml | yq '{"store": .data.OPENFGA_STORE_ID, "authorizationModel": .data.OPENFGA_AUTHORIZATION_MODEL_ID}'

store: 01HW5SHYB4ZTFBNSH05R0NVWX8
authorizationModel: 01HW5SHYBH9EXMARE01E1MC14A
```

`make dev` will take care of spinning up containers leveraging `skaffold run` (
this includes the admin-ui backend).

To debug code, you can run your local code (outside k8s) on a different
port than the one used by the Skaffold deployed instances (so PORT != 8000) and
use the environment variables to debug your code.

The only thing required, apart from the port change, is taking note of the URL +
ports used by Skaffold to expose PODs internal ports.

You can see that in `make dev` output, something like the following:

```shell
$ make dev

Deployments stabilized in 12.191 seconds
Port forwarding service/identity-platform-admin-ui in namespace default, remote port 80 -> http://127.0.0.1:8000
Port forwarding service/kratos-admin in namespace default, remote port 80 -> http://127.0.0.1:14434
Port forwarding service/openfga in namespace default, remote port 8080 -> http://127.0.0.1:14457
Port forwarding service/hydra-admin in namespace default, remote port 4445 -> http://127.0.0.1:14445
Port forwarding service/hydra-public in namespace default, remote port 4444 -> http://127.0.0.1:14444
Port forwarding service/kratos-public in namespace default, remote port 80 -> http://127.0.0.1:14433
Port forwarding service/oathkeeper-api in namespace default, remote port 4456 -> http://127.0.0.1:14456
WARN[0118] could not map pods to service default/kratos-courier/80: no pods match service kratos-courier/80  subtask=service/kratos-courier task=PortForward
Port forwarding service/openfga in namespace default, remote port 2112 -> http://127.0.0.1:2112
Port forwarding service/kratos-courier in namespace default, remote port 80 -> http://127.0.0.1:4503
Port forwarding service/postgresql-hl in namespace default, remote port 5432 -> http://127.0.0.1:5432
Port forwarding service/openfga in namespace default, remote port 3000 -> http://127.0.0.1:3000
Port forwarding service/openfga in namespace default, remote port 8081 -> http://127.0.0.1:8081
Port forwarding service/postgresql in namespace default, remote port 5432 -> http://127.0.0.1:5433
Press Ctrl+C to exit
```

So environment variables to use for running/debugging local admin ui code
outside k8s would look like:

```shell
AUTHORIZATION_ENABLED=true
HYDRA_ADMIN_URL=http://localhost:14445
IDP_CONFIGMAP_NAME=providers
IDP_CONFIGMAP_NAMESPACE=default
KRATOS_ADMIN_URL=http://localhost:14434
KRATOS_PUBLIC_URL=http://localhost:14433
KUBECONFIG_FILE=/home/barco/.kube/config
LOG_LEVEL=DEBUG
OATHKEEPER_PUBLIC_URL=http://localhost:14456
OPENFGA_API_HOST=localhost:14457
OPENFGA_API_SCHEME=http
OPENFGA_API_TOKEN=42
OPENFGA_AUTHORIZATION_MODEL_ID=01HW5SHYBH9EXMARE01E1MC14A
OPENFGA_STORE_ID=01HW5SHYB4ZTFBNSH05R0NVWX8
PORT=8888
RULES_CONFIGMAP_NAME=oathkeeper-rules
RULES_CONFIGMAP_NAMESPACE=default
SCHEMAS_CONFIGMAP_NAME=identity-schemas
SCHEMAS_CONFIGMAP_NAMESPACE=default
TRACING_ENABLED=false
```

## Stopping

To stop the Skaffold deployment, in case `Ctrl-C` is not enough, you may run
`skaffold delete` and wait for the workspace to be cleared.
These will make sure pods are deleted.

It doesn't always work at the first attempt, so you may need to stop and re-run
the delete command.

## Environment values

To retrieve all **actual** environment values from the config map in YAML
format, you can run:

```shell
kubectl get cm identity-platform-admin-ui -o yaml | yq '.data'
```

To retrieve the `export VAR1=VALUE1 VAR2=VALUE2...` format to use when running
admin ui locally, run:

```shell
$ printf "export " ; k get cm identity-platform-admin-ui -o yaml | yq .data | tr -d '"' | tr -d " " | tr ":" "=" | while read line; do echo "$line \\ " ; done ; echo
export AUTHORIZATION_ENABLED=true \
HYDRA_ADMIN_URL=http=//hydra-admin.default.svc.cluster.local=4445 \
IDP_CONFIGMAP_NAME=providers \
IDP_CONFIGMAP_NAMESPACE=default \
KRATOS_ADMIN_URL=http=//kratos-public.default.svc.cluster.local \
KRATOS_PUBLIC_URL=http=//kratos-public.default.svc.cluster.local \
LOG_LEVEL=DEBUG \
OATHKEEPER_PUBLIC_URL=http=//oathkeeper-api.default.svc.cluster.local=4456 \
OPENFGA_API_HOST=openfga.default.svc.cluster.local=8080 \
OPENFGA_API_SCHEME=http \
OPENFGA_API_TOKEN=42 \
OPENFGA_AUTHORIZATION_MODEL_ID=01HW5SHYBH9EXMARE01E1MC14A \
OPENFGA_STORE_ID=01HW5SHYB4ZTFBNSH05R0NVWX8 \
PORT=8000 \
RULES_CONFIGMAP_FILE_NAME=access-rules.json \
RULES_CONFIGMAP_NAME=oathkeeper-rules \
RULES_CONFIGMAP_NAMESPACE=default \
SCHEMAS_CONFIGMAP_NAME=identity-schemas \
SCHEMAS_CONFIGMAP_NAMESPACE=default \
TRACING_ENABLED=false
```


## OIDC authentication

For development purposes we are going to be targeting `iam.dev.canonical.com` and use it as our OIDC provider

the credentials are available in the team LastPass account and need to be swapped in the [configmap.yaml](https://github.com/canonical/identity-platform-admin-ui/blob/main/deployments/kubectl/configMap.yaml#L29-L30) file

dev setup will rely on the application being exposed on `localhost:8000` without any path rewrite, if that is not the case the OIDC flow won't work due to how the redirect uri don't match with the expectations of the OAUTH2 client created in `iam.dev.canonical.com` (and also the `OAUTH2_REDIRECT_URI` variable in the configmap)


in the remote case we need to create a new client the process to issue it will be described in the wiki together with a stop gap solution using other components (namely Dex)
