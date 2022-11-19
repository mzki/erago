# build and copy to shared dir.

set -eu

PRODUCT="erago-model-v2.aar"

BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"
TARGET_FLAGS=android # can use "android/arm,android/amd64" to shrink data size

echo "building ${PRODUCT}..."
# needs -androidapi 19 to use modern android SDK and NDK
# See https://github.com/golang/go/issues/52470.
gomobile bind ${BUILD_FLAGS} -androidapi 19 -target ${TARGET_FLAGS} -o ${PRODUCT} . || exit 1
