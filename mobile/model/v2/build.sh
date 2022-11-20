# build and copy to shared dir.

set -eu

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
BUILD_FLAGS="-gcflags=-trimpath=${GOPATH}"
TARGET_FLAGS=android # can use "android/arm,android/amd64" to shrink data size
# build mobile library
echo "building ${OUTPUT}..."
# needs -androidapi 19 to use modern android SDK and NDK
# See https://github.com/golang/go/issues/52470.
# TODO: Embed build version into binary like -x main.commit_hash=${COMMIT_HASH}?
gomobile bind ${BUILD_FLAGS} -androidapi 19 -target ${TARGET_FLAGS} -o ${OUTPUT} . || exit 1
