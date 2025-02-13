# Chronicler
An extensible tool for saving content from social networks and other sources.

The tool tries to detect to which of the specialized adapters can fetch that piece of information or
just downloads it as a web page if no adapter was found.

## Building

You need make, protobuf, go compiler and protobuf go generator:
* Arch linux: ```pacman -S go make protobuf``` and install protoc-gen go with ```go install google.golang.org/protobuf/cmd/protoc-gen-go@latest``` [protoc-gen-go](https://aur.archlinux.org/packages/protoc-gen-go) AUR package.

* Alpine: ```apk add --no-cache go make protobuf-dev``` and protoc-gen-go with command ```go install google.golang.org/protobuf/cmd/protoc-gen-go@latest```

* Ubuntu: ```apt add golang make protoc-gen-go```, no other steps needed.

And then just ```make build```

## Using

* ./main save "http://some/url" to save
* ./main view "http://some/url" to view saved url as padded text

It will save results to the ```./data/{SOME_UUID}``` directory

## Structure

### Adapters

Adapter is a module that parses some site or social network into a protobuf-defined format, extracts content, links, comment-reply structure, etc.

```
TODO: Add note about adapters
```

### Storage

Storage is an interface to put/get data "somewhere" and list what is available to get:

```go
type Storage interface {
	Put(put *PutRequest) (io.WriteCloser, error)
	Get(get *GetRequest) (io.ReadCloser, error)
	List(list *ListRequest) (*ListResponse, error)
}
```

#### Local storage

At this moment, that is only available one storage. It saves files locally with a filesystem-friendly names, so ```https://somewebsite.com/moredata/123/what``` turns into ```https___somewebsite.com_moredata_123_what```.

Typical storage folder structure:
```
.metadata/
    mapping.json
    snapshot.json_0001
    snapshot.json_0002
snapshot.json
http___somewebsite.com_more_1
http___somewebsite.com_more_2
```
### Export/View