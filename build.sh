set -Eeuo pipefail

# Configure go package and util directory if needed
export GOPATH="${PWD}"
export PATH="$PATH:$GOPATH/bin/"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

cd $GOPATH/src/main/
go install
cd ..
