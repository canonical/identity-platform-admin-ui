# Admin UI

This is the Admin UI for the Canonical identity platform.

## Development server

Haproxy needs to be installed:

```shell
apt install haproxy
```

and configured `/etc/haproxy/haproxy.cfg` with the following:

```text
global
  daemon

defaults
  mode  http

frontend iam_frontend
  bind 172.17.0.1:8000
  default_backend iam_admin_be

backend iam_admin_be
  server admin_be 127.0.0.1:8000
```

and restart it:

```shell
service haproxy restart
```

Start the build server as described in the main README.md in the root of the
repo:

```shell
make dev
```

Install `dotrun` as described
in [HERE](https://github.com/canonical/dotrun#installation).
Launch it from
the `ui/` directory of the repo:

```shell
dotrun
```

Browse <http://localhost:8411/> to reach IAM Admin UI.
