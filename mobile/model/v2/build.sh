# build and copy to shared dir.

set -eux

OUTPUTDIR=./

while getopts d: opt; do
	case "$opt" in
		d)
			OUTPUTDIR="$OPTARG"
			;;
		\?)
			echo "Usage: build.sh [-d OUTPUTDIR]"
			exit 1
			;;
	esac
done
shift $((OPTIND - 1))

# build version
VERSION=`git describe --tags --abbrev=0`
VERSION_FOR_FILE=`echo "$VERSION" | sed -e "s/\./_/g"`
COMMIT_HASH=`git rev-parse --short HEAD`
# output path
PRODUCT="erago_${VERSION_FOR_FILE}_android_model-v2.aar"
OUTPUT=$OUTPUTDIR/$PRODUCT
# flags
LD_FLAGS="-s -w -X github.com/mzki/erago/infra/buildinfo.version=$VERSION -X github.com/mzki/erago/infra/buildinfo.commitHash=$COMMIT_HASH"
BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"
TARGET_FLAGS=android # can use "android/arm,android/amd64" to shrink data size
# build mobile library
echo "building ${OUTPUT}..."
# needs -androidapi XX (e.g. 19) to use modern android SDK and NDK
# See https://github.com/golang/go/issues/52470.
# The minimum api level is increased according to NDK version in the builder image, 
# such as github action runner image.
# See https://github.com/actions/runner-images for runner image details.
# TODO: Embed build version into binary like -x main.commit_hash=${COMMIT_HASH}?
gomobile bind ${BUILD_FLAGS} -ldflags="${LD_FLAGS}" -androidapi 21 -target ${TARGET_FLAGS} -o ${OUTPUT} . || exit 1
