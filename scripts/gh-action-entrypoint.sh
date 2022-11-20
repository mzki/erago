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

#echo $*
/bin/bash -c "$*"