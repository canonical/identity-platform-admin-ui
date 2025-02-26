## Test server POC

This is a tiny poc to prove that the work done on [
`identity-platform-api`](https://github.com/canonical/identity-platform-api) is actually working.

Without building the whole app, you can simply run:

```shell
  go run entrypoint.go
```

You will then have a mock Roles API v0 available at `localhost:8081`.

Results will always be empty and simple logs will appear in the console, proving the import from the newly created
module
is actually working.

Watch out for [**PR 4**](https://github.com/canonical/identity-platform-api/pull/4), once that is merged, the import
path will change:

```shell
  "github.com/canonical/identity-platform-api/gen/roles"  => "github.com/canonical/identity-platform-api/v0/roles"
```
