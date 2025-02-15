package main

import (
	"net/http"
	"os"

	"chronicler/adapter"
	"chronicler/adapter/fourchan"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/resolver"
	"chronicler/viewer"
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

	twitterToken := os.Getenv("TWITTER_TOKEN")

	r := resolver.NewResolver(
		root,
		common.NewHttpDownloader(httpClient),
		[]adapter.Adapter{
			twitter.NewAdapter(twitter.NewClient(httpClient, twitterToken)),
			fourchan.NewAdapter(httpClient),
			pikabu.NewAdapter(httpClient),
			reddit.NewAdapter(httpClient),
			web.NewAdapter(httpClient),
		},
	)
	r.Start()
	r.Resolve(&opb.Link{Href: args[0]})
	r.Wait()
	r.Stop()
}
