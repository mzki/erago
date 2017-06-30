# build and copy to shared dir.

DST=/mnt/LM18share/
PRODUCT="model.aar"

BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"

echo "building mobile.aar..."
gomobile bind ${BUILD_FLAGS} -target android local/erago/mobile/model || exit 1
echo "copy ${PRODUCT} to ${DST}"
cp ${PRODUCT} ${DST} || exit 1
