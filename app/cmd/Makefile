# build version
VERSION          := $(shell git describe --tags --abbrev=0)
VERSION_FOR_FILE := $(shell echo "$(VERSION)" | sed -e "s/\./_/g")
COMMIT_HASH      := $(shell git rev-parse --short HEAD)

# target and source
SOURCENAME=main.go
BINARYNAME_X11=erago_$(VERSION_FOR_FILE)_linux_amd64
BINARYNAME_MAC=erago_$(VERSION_FOR_FILE)_darwin_amd64
BINARYNAME_WIN=erago_$(VERSION_FOR_FILE)_windows_amd64.exe
BINARYNAME_WIN_CLI=erago_$(VERSION_FOR_FILE)_windows_amd64_cli.exe

OUTPUTDIR ?= .

# build flags
GOBUILD=go build

GOLDFLAGS =-s -w -X github.com/mzki/erago/infra/buildinfo.version=$(VERSION) -X github.com/mzki/erago/infra/buildinfo.commitHash=$(COMMIT_HASH)
GOGCFLAGS =-trimpath=$(GOPATH)
GOASMFLAGS=-trimpath=$(GOPATH)

GOLDFLAGS_WIN=$(GOLDFLAGS) -H windowsgui # disable launching console on windows
GOLDFLAGS_WIN_CLI=$(GOLDFLAGS)

GOOPTION_X11 = -ldflags="$(GOLDFLAGS)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"
GOOPTION_MAC = -ldflags="$(GOLDFLAGS)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"
GOOPTION_WIN = -ldflags="$(GOLDFLAGS_WIN)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"
GOOPTION_WIN_CLI = -ldflags="$(GOLDFLAGS_WIN_CLI)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"

# environment
X11ENV=CC=gcc CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOROOT_FINAL="GOROOT" GO111MODULE=on
MACENV=CC=o64-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 GOROOT_FINAL="GOROOT" GO111MODULE=on
WINENV=CC=gcc CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOROOT_FINAL="GOROOT" GO111MODULE=on


all: binary

linux: $(SOURCENAME)
	$(X11ENV) $(GOBUILD) $(GOOPTION_X11) -o $(OUTPUTDIR)/$(BINARYNAME_X11)

darwin: $(SOURCENAME)
	$(MACENV) $(GOBUILD) $(GOOPTION_MAC) -o $(OUTPUTDIR)/$(BINARYNAME_MAC)

windows: $(SOURCENAME)
	$(WINENV) $(GOBUILD) $(GOOPTION_WIN) -o $(OUTPUTDIR)/$(BINARYNAME_WIN)

windows_cli: $(SOURCENAME)
	$(WINENV) $(GOBUILD) $(GOOPTION_WIN_CLI) -o $(OUTPUTDIR)/$(BINARYNAME_WIN_CLI)

binary: $(SOURCENAME)
	$(GOBUILD) 

.PHONY: clean
clean:
	$(RM) $(BINARYNAME_X11) $(BINARYNAME_WIN) $(BINARYNAME_MAC)
