package main

import (
	"chronicler/adapter"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/web"
	"chronicler/common"
	"chronicler/resolver"
	"net/http"
	"os"

	opb "chronicler/proto"
	"chronicler/viewer"
	"sync"
)

const (
	root = "data"
)

func main() {
	switch os.Args[1] {
	case "save":
		save(os.Args[2:])
	case "view":
		view(os.Args[2:])
	}
}

func view(args []string) {
	(&viewer.Viewer{
		Root: root,
	}).View(common.UUID4For(&opb.Link{Href: args[0]}))
}

func save(args []string) {
	httpClient := &http.Client{}
	r := resolver.NewResolver(
		root,
		common.NewHttpDownloader(httpClient),
		[]adapter.Adapter{
			pikabu.NewAdapter(httpClient),
			web.NewAdapter(httpClient),
		},
	)
	r.Resolve(&opb.Link{Href: args[0]})

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
