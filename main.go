package main

import (
	"chronicler/adapter/pikabu"
	"chronicler/common"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	opb "chronicler/proto"
	"chronicler/storage"
)

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

func main() {
	root := "data"

	logger := common.NewLogger("Main")
	httpClient := &http.Client{}
	loader := common.NewHttpDownloader(httpClient)
	ad := pikabu.NewAdapter(httpClient)
	link := &opb.Link{Href: os.Args[1]}

	id := common.UUID4For(link)
	baseStorage, err := storage.NewLocalStorage(filepath.Join(root, id))
	if err != nil {
		os.Exit(-1)
	}
	s := storage.BlockStorage{Storage: baseStorage}

	objs, err := ad.Get(link)
	if err != nil {
		logger.Errorf("Cannot download/parse link: %s", err)
		os.Exit(-1)
	}

	bytesWritten, err := s.PutJson(&storage.PutRequest{Url: "objects.json"}, objs)
	if err != nil {
		logger.Errorf("Cannot convert result to json: %s", err)
		os.Exit(-1)
	}
	logger.Infof("Saved objects.json, written bytes: %d", bytesWritten)

	filesToLoad := map[*url.URL]bool{}
	for _, obj := range objs {
		for _, attachment := range obj.Attachment {
			if attachment.Mime == "" {
				continue
			}
			fileUrl, err := url.Parse(attachment.Url)
			if err != nil {
				logger.Warningf("Cannot parse url \"%s\" from object %s: %s", obj.Id, fileUrl, err)
				continue
			}
			filesToLoad[fileUrl] = true
		}
	}

	logger.Infof("Files to download: %d", len(filesToLoad))
	for k := range filesToLoad {
		targetPath := filepath.Join(id, sanitizeUrl(k.Path))
		size, err := loader.Download(k.String(), s)
		if err != nil {
			logger.Warningf("Cannot download %s to %s: %s", k, targetPath, err)
			continue
		}
		if size == -1 {
		} else {
			logger.Infof("Downloaded %s to %s, sise %d", k, targetPath, size)
		}
	}
	logger.Infof("Saved objects: %d, files: %d", len(objs), len(filesToLoad))
}
