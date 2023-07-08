#!/bin/bash

# Usage:
#   gh-action-entrypoint.sh <launch-script-path> <args...>
#
# This script is a proxy of the main shell scripts from github actions.
# The porpose is to substitude environment variables before passing arguments to main script
# which user wants to launch.
#
# See https://docs.github.com/en/actions/creating-actions/dockerfile-support-for-github-actions#entrypoint. 

set -eu

# Need to configure safe.directory to perform git command inside user specific container
# See https://zenn.dev/gorohash/articles/72c9c84778194f
# See https://medium.com/@janloo/github-actions-detected-dubious-ownership-in-repository-at-github-workspace-how-to-fix-b9cc127d4c04
sh -c "git config --global --add safe.directory $PWD"

#echo $*
/bin/bash -c "$*"