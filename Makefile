SHELL = /bin/sh
GO_OUT = "."

deps:
	$(MAKE) -C proto all

go.mod:
	go mod init chronicler
	go mod tidy

build: deps go.mod
	go build -o main main.go

run: build
	./main

clean:
	$(MAKE) -C proto clean
	rm -rfv main chronicler go.mod go.sum
