# ReBAC Admin Backend

This package is a thin library that implements the *ReBAC Admin UI* OpenAPI [spec][openapi-spec]. With this library comes a set of abstractions that product services, like JAAS or Identity Platform Admin UI, need to implement and plugin. The library itself does not directly communicate with the underlying authorization provider (i.e., OpenFGA).

[openapi-spec]: https://github.com/canonical/openfga-admin-openapi-spec

## Development

Here are some Makefile targets that could help with the development:

- `pull-spec` to pull the latest stable OpenAPI spec and generate the request/response types accordingly.
