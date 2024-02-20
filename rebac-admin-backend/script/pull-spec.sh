#!/usr/bin/env sh

# Generate Go server and types from an OpenAPI spec definition file.
# A local file path can be passed as the first parameter to the script.
#
# If no parameter is passed, the OpenAPI spec will be downloaded from
# the Github repository canonical/openfga-admin-openapi-spec.
#
# This script should be run from the Makefile's parent directory.

set -e

OPENAPI_SPEC_FILE="$1"

if [ -z "$OPENAPI_SPEC_FILE" ]; then
  _tmpdir=/tmp/openfga-admin-openapi-spec

  cleanup() {
    rm -rf "$_tmpdir"
  }
  trap cleanup 0

  cleanup

  mkdir -p "$_tmpdir"
  git clone -q --depth=1 git@github.com:canonical/openfga-admin-openapi-spec "$_tmpdir"
  OPENAPI_SPEC_FILE="${_tmpdir}/openapi.yaml"

elif ! [ -r "$OPENAPI_SPEC_FILE" ]; then
  echo "Can't read the '$OPENAPI_SPEC_FILE', exiting";
  exit 1
fi

go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -generate types,spec -package resources "$OPENAPI_SPEC_FILE" > v1/resources/generated_types.go
oapi-codegen -generate chi-server -package resources "$OPENAPI_SPEC_FILE" > v1/resources/generated_server.go
