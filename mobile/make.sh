# build and copy to shared dir.

set -e

DST=/mnt/LMDE2share/

GOOPTION="-gcflags=\"-trimpath=${GOPATH}\""


# set final environment
export GOROOT_FINAL="GOROOT"

# build *.aar file.
echo "building mobile.aar..."
gomobile bind ${GOOPTION} -target android local/erago/mobile
# echo "copy eragoj.aar to ${DST}"
# cp ./eragoj.aar ${DST} || exit 1
