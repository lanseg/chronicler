set -Eeuo pipefail

# Configure go package and util directory if needed
export GOPATH="${PWD}"
export PATH="$PATH:$GOPATH/bin/"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc --proto_path=src --go_out=src/ telegram/telegram.proto

cd $GOPATH/src/main/
go install
cd ..
