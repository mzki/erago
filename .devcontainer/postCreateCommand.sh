# protobuf compiler installation: See https://protobuf.dev/installation/
PB_REL="https://github.com/protocolbuffers/protobuf/releases"
curl -LO $PB_REL/download/v30.2/protoc-30.2-linux-x86_64.zip
unzip protoc-30.2-linux-x86_64.zip -d $HOME/.local
export PATH="$PATH:$HOME/.local/bin"
rm protoc-30.2-linux-x86_64.zip

# install go-plugin for protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# ensure working go executable
go version
