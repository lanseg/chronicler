package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
)

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
