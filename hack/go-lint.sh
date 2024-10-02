#!/bin/bash
# If CI is true, run golangci-lint directly. If we're on a developer's
# local machine, run golangci-lint from a container so we're ensuring
# a consistent environment.

set -ex

if [ "$OPENSHIFT_CI" = "true" ];
then
  golangci-lint "${@}"
else
  DOCKER=${DOCKER:-podman}

  if ! which "$DOCKER" > /dev/null 2>&1;
  then
    echo "$DOCKER not found, please install."
    exit 1
  fi

  # Check if running on Linux
  VOLUME_OPTION=""
  if [[ "$(uname -s)" == "Linux" ]]; then
    VOLUME_OPTION=":z"
  fi

  $DOCKER run --rm \
    --volume "${PWD}:/go/src/github.com/openshift-eng/openshift-tests-extension${VOLUME_OPTION}" \
    --workdir /go/src/github.com/openshift-eng/openshift-tests-extension \
    docker.io/golangci/golangci-lint:v1.54.2 \
    golangci-lint "${@}"
fi
