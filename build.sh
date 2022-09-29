set -Eeuo pipefail

export ROOT="$PWD"
export PATH="$PATH:$HOME/go/bin"

cleanup() {
  find -iname '*.sum' -delete
  find -iname '*.pb.go' -delete
  
}

cleanup
go mod tidy
go build -o bin/chronist chronist/main.go

