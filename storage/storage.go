package storage

import (
	"crypto/sha512"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	rpb "chronicler/proto/records"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"
)

const (
	recordsetFileName = "record.json"
)

func hashSource(src *rpb.Source) string {
	if src == nil {
		return ""
	}
	checksum := []byte{}
	checksum = append(checksum, []byte(src.SenderId)...)
	checksum = append(checksum, []byte(src.ChannelId)...)
	checksum = append(checksum, []byte(src.MessageId)...)
	checksum = append(checksum, []byte(src.Url)...)
	checksum = append(checksum, byte(src.Type))
	return fmt.Sprintf("%x", sha512.Sum512(checksum))
}

func getRecordSetId(set *rpb.RecordSet) string {
	if set.Id != "" {
		return set.Id
	}
	if set.Request == nil {
		return ""
	}
	if set.Request.Parent != nil {
		return hashSource(set.Request.Parent)
	}
	if set.Request.Source != nil {
		return hashSource(set.Request.Source)
	}
	return ""
}

func getRecordId(record *rpb.Record) string {
	return fmt.Sprintf("%x", sha512.Sum512(
		[]byte(hashSource(record.Source)+hashSource(record.Parent))))
}

func mergeFiles(base []*rpb.File, other []*rpb.File) []*rpb.File {
	result := []*rpb.File{}
	filesById := collections.GroupBy(append(base, other...), func(f *rpb.File) string {
		return f.FileId
	})
	for _, group := range filesById {
		baseFile := group[0]
		for _, more := range group[1:] {
			if more.FileUrl != "" {
				baseFile.FileUrl = more.FileUrl
			}
		}
		result = append(result, baseFile)
	}
	return result
}

func mergeRecords(base []*rpb.Record, other []*rpb.Record) []*rpb.Record {
	if base == nil && other == nil {
		return []*rpb.Record{}
	} else if base == nil {
		return append(other)
	} else if other == nil {
		return append(base)
	}
	result := []*rpb.Record{}
	distinctRecords := collections.NewSet(other)
	distinctRecords.AddSet(collections.NewSet(other))
	recordsById := collections.GroupBy(distinctRecords.Values(), getRecordId)
	for _, group := range recordsById {
		baseRecord := group[0]
		for _, more := range group[1:] {
			baseRecord.Links = collections.Unique[string](append(baseRecord.Links, more.Links...))
			sort.Strings(baseRecord.Links)

			baseRecord.Files = mergeFiles(baseRecord.Files, more.Files)
			sort.Slice(baseRecord.Files, func(i int, j int) bool {
				return baseRecord.Files[i].FileId < baseRecord.Files[j].FileId
			})
		}
		result = append(result, baseRecord)
	}
	return result
}

func mergeUserMetadata(base []*rpb.UserMetadata, other []*rpb.UserMetadata) []*rpb.UserMetadata {
	result := []*rpb.UserMetadata{}
	distinctRecords := collections.NewSet(base)
	distinctRecords.AddSet(collections.NewSet(other))
	recordsById := collections.GroupBy(distinctRecords.Values(), func(u *rpb.UserMetadata) string {
		return u.GetId()
	})
	for _, group := range recordsById {
		baseRecord := group[0]
		for _, more := range group[1:] {
			baseRecord.Quotes = collections.Unique[string](append(baseRecord.Quotes, more.Quotes...))
			sort.Strings(baseRecord.Quotes)
			if baseRecord.Username == "" {
				baseRecord.Username = more.Username
			}
		}
		result = append(result, baseRecord)
	}
	return result
}

type Storage interface {
	SaveRecords(r *rpb.RecordSet) error
	ListRecords() optional.Optional[[]*rpb.RecordSet]
	GetFile(id string, filename string) optional.Optional[[]byte]
}

type LocalStorage struct {
	Storage

	downloader *HttpDownloader
	fs         *RelativeFS

	recordCache map[string]string
	logger      *util.Logger
	root        string
}

func (s *LocalStorage) refreshCache() error {
	s.logger.Debugf("Refreshing record cache")
	s.recordCache = map[string]string{}
	files, err := s.fs.ListFiles("")
	if err != nil {
		return err
	}
	for _, info := range files {
		ReadJSON[rpb.RecordSet](s.fs, filepath.Join(info.Name(), recordsetFileName)).
			IfPresent(func(record *rpb.RecordSet) {
				id := record.Id
				if id == "" {
					id = getRecordSetId(record)
				}
				s.recordCache[id] = info.Name()
			})
	}
	return nil
}

func (s *LocalStorage) writeRecordSet(r *rpb.RecordSet) error {
	root := getRecordSetId(r)
	s.logger.Debugf("Saving record to %s", root)
	if err := s.fs.MkDir(root); err != nil {
		return err
	}
	if err := s.fs.WriteJSON(filepath.Join(root, recordsetFileName), r); err != nil {
		return err
	}
	for _, r := range r.Records {
		for _, link := range r.Links {
			if util.IsYoutubeLink(link) && strings.Contains(link, "v=") {
				s.logger.Debugf("Found youtube link: %s", link)
				if err := util.DownloadYoutube(link, s.fs.Resolve(root)); err != nil {
					s.logger.Warningf("Failed to download youtube video: %s", err)
				}
			}
		}

		for _, file := range r.GetFiles() {
			fileUrl, err := url.Parse(file.GetFileUrl())
			if err != nil || fileUrl.String() == "" {
				s.logger.Warningf("Malformed url for file: %s", file)
				continue
			}
			fname := path.Base(fileUrl.Path)
			if err := s.downloader.Download(file.GetFileUrl(), s.fs.Resolve(filepath.Join(root, fname))); err != nil {
				s.logger.Warningf("Failed to download file: %s: %s", file, err)
			}
		}

	}
	s.logger.Infof("Saved new record to %s", root)
	return nil
}

func (s *LocalStorage) GetFile(id string, filename string) optional.Optional[[]byte] {
	return s.fs.Read(filepath.Join(id, filename))
}

func (s *LocalStorage) ListRecords() optional.Optional[[]*rpb.RecordSet] {
	s.refreshCache()
	result := []*rpb.RecordSet{}
	for _, path := range s.recordCache {
		ReadJSON[rpb.RecordSet](s.fs, filepath.Join(path, recordsetFileName)).
			IfPresent(func(rs *rpb.RecordSet) {
				if rs.Id == "" {
					rs.Id = getRecordSetId(rs)
				}
				result = append(result, rs)
			})
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Request.String() < result[j].Request.String()
	})
	return optional.Of(result)
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	id := getRecordSetId(r)
	base := ReadJSON[rpb.RecordSet](s.fs, filepath.Join(id, recordsetFileName)).
		OrElse(&rpb.RecordSet{
			Id:      r.Id,
			Request: r.Request,
		})
	base.Records = mergeRecords(base.Records, r.Records)
	base.UserMetadata = mergeUserMetadata(base.UserMetadata, r.UserMetadata)
	return s.writeRecordSet(base)
}

func NewStorage(root string) Storage {
	log := util.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)
	return &LocalStorage{
		root:        root,
		logger:      log,
		fs:          NewRelativeFS(root),
		downloader:  NewHttpDownloader(nil),
		recordCache: map[string]string{},
	}
}
