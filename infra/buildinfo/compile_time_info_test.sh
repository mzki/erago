set -ux
GO_LDFLAG="-X github.com/mzki/erago/infra/buildinfo.version=v1.0.0 -X github.com/mzki/erago/infra/buildinfo.commitHash=34567#"
GOTEST_BUILDINFO_COMPILE_TIME_INFO=true go test -ldflags="$GO_LDFLAG" .