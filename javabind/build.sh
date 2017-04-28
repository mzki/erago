# build and copy to shared dir.

DST=/mnt/LMDE2share/

BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"

echo "building eragoj.aar..."
gomobile bind ${BUILD_FLAGS} -target android local/erago/javabind || exit 1
echo "copy eragoj.aar to ${DST}"
cp ./eragoj.aar ${DST} || exit 1
