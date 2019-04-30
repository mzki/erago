#!/bin/bash
set -eu

docker run -v $GOPATH/src:/go/src -w /go/src/github.com/mzki/erago/mobile/model -t gomobile-bind /bin/bash build.sh
