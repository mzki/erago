# build and copy to shared dir.

set -eu

PRODUCT="erago-model.aar"

BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"
TARGET_FLAGS=android # can use "android/arm,android/amd64" to shrink data size

echo "building ${PRODUCT}..."
gomobile bind ${BUILD_FLAGS} -target ${TARGET_FLAGS} -o ${PRODUCT} github.com/mzki/erago/mobile/model || exit 1
