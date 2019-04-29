# build version
VERSION          := $(shell git describe --tags --abbrev=0)
VERSION_FOR_FILE := $(shell echo "$(VERSION)" | sed -e "s/\./_/g")
COMMIT_HASH      := $(shell git rev-parse --short HEAD)

# target and source
SOURCENAME=main.go
BINARYNAME_X11=erago_$(VERSION_FOR_FILE)_linux_amd64
BINARYNAME_WIN=erago_$(VERSION_FOR_FILE)_windows_amd64.exe

# build flags
GOBUILD=go build

GOLDFLAGS =-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH)
GOGCFLAGS =-trimpath=$(GOPATH)
GOASMFLAGS=-trimpath=$(GOPATH)

GOLDFLAGS_WIN=$(GOLDFLAGS) -H windowsgui # disable launching console on windows

GOOPTION_X11 = -ldflags="$(GOLDFLAGS)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"
GOOPTION_WIN = -ldflags="$(GOLDFLAGS_WIN)" -gcflags="$(GOGCFLAGS)" -asmflags="$(GOASMFLAGS)"

# environment
X11ENV=GOOS=linux GOARCH=amd64 GOROOT_FINAL="GOROOT"
WINENV=GOOS=windows GOARCH=386 GOROOT_FINAL="GOROOT"


all: $(BINARYNAME_X11) $(BINARYNAME_WIN)

linux: $(SOURCENAME)
	$(X11ENV) $(GOBUILD) $(GOOPTION_X11) -o $(BINARYNAME_X11)

windows: $(SOURCENAME)
	$(WINENV) $(GOBUILD) $(GOOPTION_WIN) -o $(BINARYNAME_WIN)

.PHONY: clean
clean:
	$(RM) $(BINARYNAME_X11) $(BINARYNAME_WIN)
