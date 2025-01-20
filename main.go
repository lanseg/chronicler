package main

import (
	"chronicler/adapter/pikabu"
	"chronicler/common"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	opb "chronicler/proto"
)

func main() {
	logger := common.NewLogger("Main")
	link := &opb.Link{Href: "12267205"}
	httpClient := &http.Client{}
	loader := common.NewHttpDownloader(httpClient)
	ad := pikabu.NewAdapter(httpClient)

	id := common.UUID4For(link)
	if err := os.Mkdir(id, 0777); err != nil {
		if errors.Is(err, os.ErrExist) {
			logger.Infof("Folder already exist, updating %s", id)
		} else {
			logger.Errorf("Cannot create folder for saving: %s", err)
			os.Exit(-1)
		}
	}

	objs, err := ad.Get(link)
	if err != nil {
		logger.Errorf("Cannot download/parse link: %s", err)
		os.Exit(-1)
	}

	str, err := json.Marshal(objs)
	if err != nil {
		logger.Errorf("Cannot convert result to json: %s", err)
		os.Exit(-1)
	}

	resultJsonFile := filepath.Join(id, "objects.json")
	if err = os.WriteFile(resultJsonFile, str, 0666); err != nil {
		logger.Errorf("Cannot save result to json file %s: %s", resultJsonFile, err)
		os.Exit(-1)
	}

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
		path := strings.Split(k.Path, "/")
		targetPath := filepath.Join(id, path[len(path)-1])
		size, err := loader.Download(k.String(), targetPath)
		if err != nil {
			logger.Warningf("Cannot download %s to %s: %s", k, targetPath, err)
			continue
		}
		logger.Infof("Downloaded %s to %s, sise %d", k, targetPath, size)
	}
}
