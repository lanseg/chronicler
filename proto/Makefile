SHELL = /bin/sh
GO_OUT = "./"

protoc-go: objects.proto
	protoc -I=. --go_out=$(GO_OUT) --go_opt=paths=source_relative objects.proto

all: protoc-go

clean:
	rm -rfv objects.pb.go
