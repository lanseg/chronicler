package exporter

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chronicler/common"
	"chronicler/storage"

	opb "chronicler/proto"
)

const objectFileName = "snapshot.json"

func HtmlExporter(store storage.Storage) Exporter {
	return &htmlExporter{
		store:  store,
		logger: common.NewLogger("HtmlExporter"),
	}
}

type htmlExporter struct {
	Exporter

	logger *common.Logger
	store  storage.Storage
}

func (le *htmlExporter) exportFile(target string, href string, typeName string, reader io.Reader) (int64, error) {
	fileTarget := filepath.Join(target, sanitizeURL(href, typeName))
	if _, err := os.Stat(fileTarget); err == nil {
		le.logger.Infof("Skipped, %q: already exists", fileTarget)
		return 0, nil
	}
	baseDir := filepath.Dir(fileTarget)
	if baseDir != "" {
		os.MkdirAll(baseDir, os.ModePerm)
	}
	f, err := os.OpenFile(fileTarget, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, reader)
}

func (le *htmlExporter) Export(destination string) error {
	le.logger.Infof("Exporting to %q", destination)
	bs := storage.BlockStorage{Storage: le.store}

	le.logger.Infof("Loading snapshot from storage: %q", objectFileName)
	result := &opb.Snapshot{}
	err := bs.GetJSON(&storage.GetRequest{Url: "snapshot.json"}, result)
	if err != nil {
		le.logger.Warningf("cannot read snapshot json file %q: %q",
			"snapshot.json", err)
		err = bs.GetProto(&storage.GetRequest{Url: "snapshot.binpb"}, result)
	}
	if err != nil {
		le.logger.Warningf("cannot read snapshot binary proto file %q: %q",
			"snapshot.binpb", err)
		return err
	}

	total := len(result.Objects)
	if total == 0 {
		le.logger.Warningf("No objects found in the snapshot: %q", result)
		return nil
	}

	le.logger.Infof("Loaded objects: %d", total)
	if err := os.MkdirAll(destination, 0766); err != nil {
		return err
	}

	siteToLocal := buildMapping(result)
	le.logger.Debugf("Link mapping done, got records: %d", len(siteToLocal))

	for i, obj := range result.Objects {
		le.logger.Infof("[%06d of %06d] object exporting %q", i, total, obj.Id)
		body := strings.Builder{}
		for _, content := range obj.Content {
			body.WriteString(content.Text)
		}
		converted := []byte(convertLinks(body.String(), siteToLocal[obj.Id]))
		if _, err := le.exportFile(destination, obj.Id, http.DetectContentType(converted), bytes.NewReader(converted)); err != nil {
			le.logger.Errorf("[%06d of %06d] error exporting %q: %q", i, total, obj.Id, err)
		}
	}
	le.logger.Debugf("Exported all %d objects", len(result.Objects))

	for i, obj := range result.Objects {
		mapping := siteToLocal[obj.Id]
		for _, att := range obj.Attachment {
			attUrl := att.Url.Href
			fileTarget := filepath.Join(destination, sanitizeURL(attUrl, att.Mime))
			if _, err := os.Stat(fileTarget); err == nil {
				le.logger.Infof(
					"[%06d of %06d] attachment exporting %q to %q (skipped, already exists)",
					i, total, attUrl, fileTarget)
				continue
			}

			le.logger.Infof(
				"[%06d of %06d] attachment exporting %q to %q", i, total, attUrl, fileTarget)
			request := &storage.GetRequest{Url: attUrl}
			fileBytes, err := bs.GetBytes(request)
			if err != nil {
				le.logger.Warningf("cannot open file %q for reading: %q", attUrl, err)
				continue
			}

			var reader io.ReadCloser
			if strings.HasPrefix(att.Mime, "text") {
				reader = io.NopCloser(bytes.NewBufferString(convertLinks(string(fileBytes), mapping)))
			} else {
				reader = io.NopCloser(bytes.NewReader(fileBytes))
			}
			if _, err := le.exportFile(destination, attUrl, http.DetectContentType(fileBytes), reader); err != nil {
				le.logger.Errorf("[%06d of %06d] error attachment exporting %q to %q: %q", i, total,
					obj.Id, fileTarget, err)
			}
		}
	}
	return nil
}
