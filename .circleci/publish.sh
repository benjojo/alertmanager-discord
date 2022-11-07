#!/usr/bin/env bash
set -eo pipefail

DOCKER_IMAGE_TAG=speckle/alertmanager-discord

docker tag "${DOCKER_IMAGE_TAG}:${CIRCLE_SHA1}.${CIRCLE_BUILD_NUM}" "${DOCKER_IMAGE_TAG}:latest"

echo "${DOCKER_REG_PASS}" | docker login -u "${DOCKER_REG_USER}" --password-stdin "${DOCKER_REG_URL}"
docker push -a "${DOCKER_IMAGE_TAG}"
