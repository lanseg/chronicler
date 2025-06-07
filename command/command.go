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
	"errors"
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
	if len(args) < 2 {
		return fmt.Errorf("Export requires two arguments: <link> and <target>, but got %q", args)
	}
	storage, err := getStorage(s, args[0])
	if err != nil {
		return err
	}
	return exporter.NewLocalExporter(storage).Export(args[1])
}

func View(s *Settings, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("View requires one argument: <link>, but got none")
	}
	storage, err := getStorage(s, args[0])
	if err != nil {
		return err
	}
	return exporter.NewTextExporter(storage).Export("")
}

func List(s *Settings, args []string) error {
	logger := common.NewLogger("Command-List")
	dir, err := os.ReadDir(s.Storage.Root)
	if err != nil {
		return errors.Join(fmt.Errorf("cannot read storage dir %q", s.Storage.Root), err)
	}
	snapshots := []*opb.Snapshot{}
	for _, d := range dir {
		ls, err := storage.NewLocalStorage(filepath.Join(s.Storage.Root, d.Name()))
		if err != nil {
			logger.Warningf("cannot read storage in the folder %q: %q", d, err)
			continue
		}
		bs := storage.BlockStorage{Storage: ls}
		snapshot := &opb.Snapshot{}
		if err = bs.GetJSON(&storage.GetRequest{Url: "snapshot.json"}, snapshot); err != nil {
			logger.Warningf("cannot read snapshot json file %q in the folder %q: %q",
				"snapshot.json", d, err)
			err = bs.GetProto(&storage.GetRequest{Url: "snapshot.binpb"}, snapshot)
		}
		if err != nil {
			logger.Warningf("cannot read snapshot binary proto file %q in the folder %q: %q",
				"snapshot.binpb", d, err)
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
		fmt.Printf("%03d [%s] %s\n", i, fetchTime, snapshot.Link.Href)
	}
	return nil
}

func Save(s *Settings, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Save requires one argument: <link>, but got none")
	}
	link, err := opb.ParseLink(args[0])
	if err != nil {
		return err
	}
	logger := common.NewLogger("Save")
	httpClient := common.NewHttpClient(s.HttpSettings.CachePath, s.HttpSettings.CookieJar)
	adapters := []adapter.Adapter{}
	if s.Twitter != nil {
		adapters = append(adapters,
			twitter.NewAdapter(twitter.NewClient(httpClient, s.Twitter.Token)))
		logger.Infof("Twitter adapter loaded")
	}
	if s.Reddit != nil {
		adapters = append(adapters,
			reddit.NewAdapter(httpClient, &reddit.RedditAuth{AccessToken: s.Reddit.Token}))
		logger.Infof("Reddit adapter loaded")
	}
	adapters = append(adapters,
		fourchan.NewAdapter(httpClient),
		pikabu.NewAdapter(httpClient),
		web.NewAdapter(httpClient))
	r := resolver.NewResolver(
		s.Storage.Root,
		common.NewHttpDownloader(httpClient),
		adapters,
	)
	r.Start()
	r.Resolve(link)
	r.Wait()
	r.Stop()
	return nil
}
