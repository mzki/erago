# build and copy to shared dir.

set -e

DST=/mnt/LM18share/
PRODUCT="model.aar"

BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"
TARGET_FLAGS=android/arm # can use "android/arm,android/amd"

echo "building mobile.aar..."
gomobile bind ${BUILD_FLAGS} -target ${TARGET_FLAGS} local/erago/mobile/model || exit 1
echo "copy ${PRODUCT} to ${DST}"
cp ${PRODUCT} ${DST} || exit 1
