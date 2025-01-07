SHELL = /bin/sh
GO_OUT = "out/go"

protoc-go: objects.proto
	mkdir -p $(GO_OUT)
	protoc -I=. --go_out=$(GO_OUT) objects.proto	    

protoc: protoc-go

install: protoc
	mkdir -p $(DESTDIR)/
	cp -rfv $(GO_OUT)/* $(DESTDIR)/

clean:
	rm -rfv $(GO_OUT)
