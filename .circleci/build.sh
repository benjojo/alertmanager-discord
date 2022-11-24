#!/usr/bin/env bash
set -eo pipefail

if [[ -z "${VERSION}" ]]; then
  echo "VERSION environment variable should be set"
  exit 1
fi

DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG:-"speckle/alertmanager-discord"}"
export DOCKER_BUILDKIT=1

docker build --tag "${DOCKER_IMAGE_TAG}:${VERSION}" --build-arg="APPLICATION_VERSION=${VERSION}" --file ./Dockerfile .
