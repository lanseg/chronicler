package command

import (
	opb "chronicler/proto"
	"chronicler/storage"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func List(s *Settings, args []string) {
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
