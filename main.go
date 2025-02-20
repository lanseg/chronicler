package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"sort"
	"time"

	"chronicler/adapter"
	"chronicler/adapter/fourchan"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/resolver"
	"chronicler/storage"
	"chronicler/viewer"
)

const (
	root = "data"
)

func main() {
	switch os.Args[1] {
	case "list":
		list(os.Args[2:])
	case "save":
		save(os.Args[2:])
	case "view":
		view(os.Args[2:])
	}
}

func list(_ []string) {
	dir, err := os.ReadDir(root)
	if err != nil {
		return
	}
	snapshots := []*opb.Snapshot{}
	for _, d := range dir {
		ls, err := storage.NewLocalStorage(filepath.Join(root, d.Name()))
		if err != nil {
			//fmt.Printf("%03d %s\n", i, err)
			continue
		}
		bs := storage.BlockStorage{Storage: ls}
		snapshot := &opb.Snapshot{}
		if err = bs.GetObject(&storage.GetRequest{Url: "snapshot.json"}, snapshot); err != nil {
			//fmt.Printf("%03d %s\n", i, err)
			continue
		}
		snapshots = append(snapshots, snapshot)
	}
	sort.Slice(snapshots, func(i, j int) bool {
		sa := snapshots[i]
		sb := snapshots[j]
		if sb.FetchTime == nil {
			return false
		} else if sa.FetchTime == nil {
			return true
		}
		return sa.FetchTime.Seconds < sb.FetchTime.Seconds
	})
	for i, snapshot := range snapshots {
		fetchTime := "?"
		if snapshot.FetchTime != nil {
			fetchTime = time.Unix(snapshot.FetchTime.Seconds, 0).Format(time.DateTime)
		}
		fmt.Printf("%03d [%s] %s\n", i, fetchTime, snapshot.Link)
	}
}

func view(args []string) {
	(&viewer.Viewer{
		Root: root,
	}).View(common.UUID4For(&opb.Link{Href: args[0]}))
}

func save(args []string) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{Jar: jar}

	twitterToken := os.Getenv("TWITTER_TOKEN")
	redditToken := os.Getenv("REDDIT_TOKEN")

	r := resolver.NewResolver(
		root,
		common.NewHttpDownloader(httpClient),
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
