#!/usr/bin/env bash
set -eo pipefail

if [[ -z "${VERSION}" ]]; then
  echo "VERSION environment variable should be set"
  exit 1
fi

if [[ -z "${DOCKER_REG_PASS}" ]]; then
  echo "DOCKER_REG_PASS environment variable should be set"
  exit 1
fi

if [[ -z "${DOCKER_REG_USER}" ]]; then
  echo "DOCKER_REG_USER environment variable should be set"
  exit 1
fi

DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG:-"speckle/alertmanager-discord"}"

docker tag "${DOCKER_IMAGE_TAG}:${VERSION}" "${DOCKER_IMAGE_TAG}:latest"

echo "${DOCKER_REG_PASS}" | docker login -u "${DOCKER_REG_USER}" --password-stdin "${DOCKER_REG_URL}"
docker push -a "${DOCKER_IMAGE_TAG}"
