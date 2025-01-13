SHELL = /bin/sh
GO_OUT = "."

deps:
	$(MAKE) -C proto all

go.mod:
	go mod init chronicler

build: deps go.mod
	go mod tidy
	go build -o main main.go

test: build
	go vet
	go test ./...

run: build
	./main

clean:
	$(MAKE) -C proto clean
	rm -rfv main chronicler go.mod go.sum
