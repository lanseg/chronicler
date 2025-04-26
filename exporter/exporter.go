package exporter

import (
	"bytes"
	"chronicler/common"
	"chronicler/storage"
	"io"
	"os"
	"path/filepath"
	"strings"

	opb "chronicler/proto"
)

const objectFileName = "snapshot.json"

func convertLinks(text string, realToLocal map[string]string) string {
	kv := []string{}
	for link, localPath := range realToLocal {
		kv = append(kv,
			"\""+link+"\"", "\""+localPath+"\"",
			"'"+link+"'", "'"+localPath+"'")
	}
	return strings.NewReplacer(kv...).Replace(text)
}

func buildMapping(snap *opb.Snapshot) map[string]map[string]string {
	siteToLocal := map[string]map[string]string{}
	for _, obj := range snap.Objects {
		if siteToLocal[obj.Id] == nil {
			siteToLocal[obj.Id] = map[string]string{
				obj.Id: common.SanitizeWithExt(obj.Id, "", 255),
			}
		}
		currentMapping := siteToLocal[obj.Id]

		for _, att := range obj.Attachment {
			safeUrl := common.SanitizeWithExt(att.Url.Href, att.Mime, 255)
			for _, variant := range append(att.Url.Variants, att.Url.Href) {
				currentMapping[variant] = safeUrl
			}
		}
	}
	return siteToLocal
}

type Exporter interface {
	Export(destination string) error
}

func NewLocalExporter(store storage.Storage) Exporter {
	return &localExporter{
		store:  store,
		logger: common.NewLogger("LocalExporter"),
	}
}

type localExporter struct {
	Exporter

	logger *common.Logger
	store  storage.Storage
}

func (le *localExporter) exportFile(target string, href string, typeName string, reader io.Reader) (int64, error) {
	fileTarget := filepath.Join(target, common.SanitizeWithExt(href, typeName, 255))
	if _, err := os.Stat(fileTarget); err == nil {
		le.logger.Infof("Skipped, %q: already exists", fileTarget)
		return 0, nil
	}
	f, err := os.OpenFile(fileTarget, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, reader)
}

func (le *localExporter) Export(destination string) error {
	le.logger.Infof("Exporting to %q", destination)
	store := storage.BlockStorage{Storage: le.store}

	le.logger.Infof("Loading snapshot from storage: %q", objectFileName)
	result := &opb.Snapshot{}
	if err := store.GetObject(&storage.GetRequest{Url: objectFileName}, &result); err != nil {
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
		if _, err := le.exportFile(destination, obj.Id, "text/plain", bytes.NewReader(converted)); err != nil {
			le.logger.Errorf("[%06d of %06d] error exporting %q: %q", i, total, obj.Id, err)
		}
	}
	le.logger.Debugf("Exported all %d objects", len(result.Objects))

	for i, obj := range result.Objects {
		mapping := siteToLocal[obj.Id]
		for _, att := range obj.Attachment {
			attUrl := att.Url.Href
			fileTarget := filepath.Join(destination, common.SanitizeWithExt(attUrl, att.Mime, 255))
			if _, err := os.Stat(fileTarget); err == nil {
				le.logger.Infof(
					"[%06d of %06d] attachment exporting %q to %q (skipped, already exists)",
					i, total, attUrl, fileTarget)
				continue
			}

			le.logger.Infof(
				"[%06d of %06d] attachment exporting %q to %q", i, total, attUrl, fileTarget)
			request := &storage.GetRequest{Url: attUrl}
			fileBytes, err := store.GetBytes(request)
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
			if _, err := le.exportFile(destination, attUrl, "text/plain", reader); err != nil {
				le.logger.Errorf("[%06d of %06d] error attachment exporting %q to %q: %q", i, total,
					obj.Id, fileTarget, err)
			}
		}
	}
	return nil
}
