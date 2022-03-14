set -Eeuo pipefail

export ROOT="$PWD"
export PATH="$PATH:$HOME/go/bin"

cleanup() {
  find -iname '*.sum' -delete
  find -iname '*.pb.go' -delete
  
}

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

cleanup
protoc --proto_path=telegram/ --go_out=telegram/ --go_opt=paths=source_relative telegram/telegram.proto
go mod tidy

