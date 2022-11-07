#!/usr/bin/env bash
set -eo pipefail

DOCKER_IMAGE_TAG=speckle/alertmanager-discord
export DOCKER_BUILDKIT=1

docker build --tag "${DOCKER_IMAGE_TAG}:${CIRCLE_SHA1}.${CIRCLE_BUILD_NUM}" --file ./Dockerfile .
