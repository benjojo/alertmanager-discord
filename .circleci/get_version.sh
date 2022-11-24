#!/bin/bash
set -eo pipefail

if [[ "${CIRCLE_TAG}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "${CIRCLE_TAG}"
    exit 0
fi

# shellcheck disable=SC2068,SC2046
LAST_RELEASE="$(git describe --always --tags $(git rev-list --tags) | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1)"
NEXT_RELEASE="$(echo "${LAST_RELEASE}" | awk -F. -v OFS=. '{$NF += 1 ; print}')"
if [[ "${CIRCLE_BRANCH}" == "main" ]]; then
    echo "${NEXT_RELEASE}-alpha.${CIRCLE_BUILD_NUM}"
    exit 0
fi

 # docker has a 128 character tag limit, so ensuring the branch name will be short enough
 # helm uses semver 2, only valid characters are a-zA-Z0-9 and hyphen '-'
# shellcheck disable=SC2034
BRANCH_NAME_TRUNCATED="$(echo "${CIRCLE_BRANCH}" | cut -c -50 | sed 's/[^a-zA-Z0-9.-]/-/g')"

echo "${NEXT_RELEASE}-branch.${BRANCH_NAME_TRUNCATED}.${CIRCLE_BUILD_NUM}"
exit 0
