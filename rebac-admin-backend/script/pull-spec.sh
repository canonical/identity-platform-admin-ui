#!/usr/bin/env sh

# Pull latest stable OpenAPI spec from GitHub and auto-generate Go types.
#
# This command should be run from the Makefile's parent directory.

set -e

_tmpdir=/tmp/openfga-admin-openapi-spec
cleanup() {
  rm -rf "$_tmpdir"
}
trap cleanup 0

cleanup
mkdir -p "$_tmpdir"
git clone -q --depth=1 git@github.com:canonical/openfga-admin-openapi-spec "$_tmpdir"

go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -generate types,spec -package resources "${_tmpdir}/openapi.yaml" > v1/resources/generated_types.go
oapi-codegen -generate chi-server -package resources "${_tmpdir}/openapi.yaml" > v1/resources/generated_server.go
