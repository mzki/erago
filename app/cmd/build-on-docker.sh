#!/bin/bash

#
# manual building with docker environment for cross compilation.
# The automated build, e.g. CI/CD,  may not be used this file.
#
set -eu

REPO_ROOT=$(dirname $(dirname ${PWD})) # 2 above directory
TARGET_ROOT=/home/gopher/src

DOCKER_IMAGE="golang-cross-local-user:1.12.5"
BUILD_TARGET=${1:- linux}

docker run \
  -v ${REPO_ROOT}:${TARGET_ROOT} \
  -w ${TARGET_ROOT}/app/cmd \
  -t $DOCKER_IMAGE \
  make $BUILD_TARGET
