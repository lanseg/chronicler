package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

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

func sanitizeUrl(remotePath string) string {
	builder := strings.Builder{}
	for _, r := range remotePath {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

type DumpHttpClient struct {
	adapter.HttpClient

	i            int
	actualClient adapter.HttpClient
}

func (dh *DumpHttpClient) Do(request *http.Request) (*http.Response, error) {
	if err := os.MkdirAll("/home/arusakov/devel/lanseg/chronicler/dump", 0777); err != nil {
		fmt.Printf("HERE-ERR-1: %s\n", err)
		return nil, err
	}
	result, err := dh.actualClient.Do(request)
	if err != nil {
		fmt.Printf("HERE-ERR-2: %s\n", err)
		return nil, err
	}
	data, err := io.ReadAll(result.Body)
	if err != nil {
		fmt.Printf("HERE-ERR-3: %s\n", err)
		return nil, err
	}
	defer result.Body.Close()
	target := sanitizeUrl(request.URL.RawQuery)
	os.WriteFile(fmt.Sprintf("dump/%s_%d", target, dh.i), data, 0777)
	dh.i += 1
	result.Body = io.NopCloser(bytes.NewReader(data))
	return result, nil
}

func save(args []string) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &DumpHttpClient{actualClient: &http.Client{Jar: jar}}

	twitterToken := os.Getenv("TWITTER_TOKEN")
	redditToken := os.Getenv("REDDIT_TOKEN")

	r := resolver.NewResolver(
		root,
		common.NewHttpDownloader(httpClient.actualClient.(*http.Client)),
		[]adapter.Adapter{
			twitter.NewAdapter(twitter.NewClient(httpClient, twitterToken)),
			fourchan.NewAdapter(httpClient),
			pikabu.NewAdapter(httpClient),
			reddit.NewAdapter(httpClient, &reddit.RedditAuth{AccessToken: redditToken}),
			web.NewAdapter(httpClient),
		},
	)
	r.Start()
	r.Resolve(&opb.Link{Href: args[0]})
	r.Wait()
	r.Stop()
}
