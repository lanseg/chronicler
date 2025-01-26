package main

import (
	"chronicler/adapter/pikabu"
	"chronicler/common"
	"chronicler/iferr"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	opb "chronicler/proto"
	"chronicler/storage"
)

func main() {
	logger := common.NewLogger("Main")
	root := "data"
	link := &opb.Link{Href: os.Args[1]}
	id := common.UUID4For(link)
	s := storage.BlockStorage{Storage: iferr.Exit(storage.NewLocalStorage(filepath.Join(root, id)))}

	httpClient := &http.Client{}
	loader := common.NewHttpDownloader(httpClient)
	ad := pikabu.NewAdapter(httpClient)

	objs := iferr.Exit(ad.Get(link))
	bytesWritten := iferr.Exit(s.PutJson(&storage.PutRequest{Url: "objects.json"}, objs))

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
		size, err := loader.Download(k.String(), s)
		if err != nil {
			logger.Warningf("Cannot download %s: %s", k, err)
			continue
		}
		if size == -1 {
			logger.Infof("No need to download file %s", k)
		} else {
			logger.Infof("Downloaded %s, size %d", k, size)
		}
	}
	logger.Infof("Saved objects: %d, files: %d", len(objs), len(filesToLoad))
}
