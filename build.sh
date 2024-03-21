#!/bin/sh

# The script requires:
# - rockcraft
# - skopeo with sudo privilege
# - yq
# - docker

set -e

rockcraft="rockcraft.yaml"
rockcraft_backup="rockcraft_bak.yaml"

restore() {
    mv "$rockcraft_backup" "$rockcraft"
    rm -f "$rockcraft_backup"
}
trap 'restore' INT TERM EXIT

# The ROCK image needs certain utilities to
# - create OpenFGA store and authorization model
# - export OpenFGA store ID and authorization model ID to Admin UI service
cp "$rockcraft" "$rockcraft_backup"
yq -i \
  '.base = "ubuntu@22.04", .parts |= ({"utils": {"plugin": "nil", "stage-packages": ["curl", "jq"]}} + .)' \
  "$rockcraft"
rockcraft pack -v

skopeo --insecure-policy \
  copy "oci-archive:identity-platform-admin-ui_$(yq -r '.version' rockcraft.yaml)_amd64.rock" \
  docker-daemon:"$IMAGE"

docker push "$IMAGE"
