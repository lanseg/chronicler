package command

import (
	"chronicler/adapter"
	"chronicler/adapter/fourchan"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
	"chronicler/common"
	"chronicler/exporter"
	opb "chronicler/proto"
	"chronicler/resolver"
	"chronicler/storage"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Command = func(*Settings, []string) error

func getStorage(s *Settings, link string) (storage.Storage, error) {
	itemLink, err := opb.ParseLink(link)
	if err != nil {
		return nil, err
	}
	storage, err := storage.NewLocalStorage(filepath.Join(s.Storage.Root, common.UUID4For(itemLink)))
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func Export(s *Settings, args []string) error {
	storage, err := getStorage(s, args[1])
	if err != nil {
		return err
	}
	return exporter.NewLocalExporter(storage).Export(args[0])
}

func View(s *Settings, args []string) error {
	storage, err := getStorage(s, args[1])
	if err != nil {
		return err
	}
	return exporter.NewTextExporter(storage).Export("")
}

func List(s *Settings, args []string) error {
	root := s.Storage.Root
	dir, err := os.ReadDir(root)
	if err != nil {
		return err
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
	return nil
}

func Save(s *Settings, args []string) error {
	httpClient := common.NewHttpClient(s.HttpSettings.CachePath, s.HttpSettings.CookieJar)
	r := resolver.NewResolver(
		s.Storage.Root,
		common.NewHttpDownloader(httpClient),
		[]adapter.Adapter{
			twitter.NewAdapter(twitter.NewClient(httpClient, s.Twitter.Token)),
			fourchan.NewAdapter(httpClient),
			pikabu.NewAdapter(httpClient),
			reddit.NewAdapter(httpClient, &reddit.RedditAuth{AccessToken: s.Reddit.Token}),
			web.NewAdapter(httpClient),
		},
	)
	r.Start()
	r.Resolve(&opb.Link{Href: args[0]})
	r.Wait()
	r.Stop()
	return nil
}
