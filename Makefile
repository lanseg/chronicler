SHELL = /bin/sh
GO_OUT = "."
COVFILE = "coverage.out"

install-hooks:
	@echo "make clean test" >> .git/hooks/pre-commit
	@chmod a+x .git/hooks/pre-commit

deps:
	$(MAKE) -C proto all

go.mod:
	go mod init chronicler

build: deps go.mod
	go mod tidy
	go build -o main main.go

test: build
	@rm -rf $(COVFILE)
	staticcheck ./...
	go vet
	go test -v -cover -coverprofile=$(COVFILE) ./...

run: build
	./main

clean:
	$(MAKE) -C proto clean
	rm -rfv main chronicler go.mod go.sum $(COVFILE)
