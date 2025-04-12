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

var (
	logger = common.NewLogger("main")
)

type Settings struct {
	Twitter *twitter.Settings `json:"twitter"`
	Reddit  *reddit.Settings  `json:"reddit"`
	Storage *storage.Settings `json:"storage"`
}

func getSettings() *Settings {
	return &Settings{
		Twitter: &twitter.Settings{
			Token: os.Getenv("TWITTER_TOKEN"),
		},
		Reddit: &reddit.Settings{
			Token: os.Getenv("REDDIT_TOKEN"),
		},
		Storage: &storage.Settings{
			Root: os.Getenv("STORAGE_ROOT"),
		},
	}
}

func getCommand() func(*Settings, []string) {
	switch os.Args[1] {
	case "list":
		return list
	case "save":
		return save
	case "view":
		return view
	case "export":
		return export
	}
	return nil
}

func main() {
	settings := getSettings()
	command := getCommand()
	if command == nil {
		logger.Infof("Unknown command (args %q)", os.Args)
		return
	}
	logger.Infof("Running command %q with args %q", os.Args[1], os.Args[2:])
	command(settings, os.Args[2:])
}

func list(s *Settings, _ []string) {
	root := s.Storage.Root
	dir, err := os.ReadDir(root)
	if err != nil {
		return
	}
	snapshots := []*opb.Snapshot{}
	for _, d := range dir {
		ls, err := storage.NewLocalStorage(filepath.Join(root, d.Name()))
		if err != nil {
			continue
		}
		bs := storage.BlockStorage{Storage: ls}
		snapshot := &opb.Snapshot{}
		if err = bs.GetObject(&storage.GetRequest{Url: "snapshot.json"}, snapshot); err != nil {
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

func view(s *Settings, args []string) {
	viewer.NewViewer(s.Storage.Root).View(common.UUID4For(&opb.Link{Href: args[0]}))
}

func export(s *Settings, args []string) {
	viewer.NewExporter(s.Storage.Root, args[1]).Export(common.UUID4For(&opb.Link{Href: args[0]}))
}

func save(s *Settings, args []string) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Minute,
	}

	twitterToken := os.Getenv("TWITTER_TOKEN")
	redditToken := os.Getenv("REDDIT_TOKEN")

	r := resolver.NewResolver(
		s.Storage.Root,
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
