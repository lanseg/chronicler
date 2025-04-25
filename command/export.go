package command

import (
	"bytes"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func convertLinks(text string, realToLocal map[string]string) string {
	kv := []string{}
	for link, localPath := range realToLocal {
		kv = append(kv, "\""+link+"\"", "\""+localPath+"\"", "'"+link+"'", "'"+localPath+"'")
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

func Export(s *Settings, args []string) {
	NewExporter(s.Storage.Root, args[0]).Export(common.UUID4For(common.OrExit(opb.ParseLink(args[1]))))
}

type Exporter struct {
	root   string
	target string
	logger *common.Logger
}

func NewExporter(root string, target string) *Exporter {
	return &Exporter{
		root:   root,
		target: target,
		logger: common.NewLogger("export"),
	}
}

func (v *Exporter) exportFile(href string, typeName string, reader io.Reader) (int64, error) {
	fileTarget := filepath.Join(v.target, common.SanitizeWithExt(href, typeName, 255))
	if _, err := os.Stat(fileTarget); err == nil {
		v.logger.Infof("Skipped, %q: already exists", fileTarget)
		return 0, nil
	}
	f, err := os.OpenFile(fileTarget, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, reader)
}

func (v *Exporter) Export(id string) error {
	v.logger.Infof("Exporting %q/%q to %q", v.root, id, v.target)
	store := storage.BlockStorage{
		Storage: common.OrExit(storage.NewLocalStorage(filepath.Join(v.root, id))),
	}

	v.logger.Infof("Loading objects from %q", filepath.Join(v.root, id, objectFileName))
	result := &opb.Snapshot{}
	if err := store.GetObject(&storage.GetRequest{Url: objectFileName}, &result); err != nil {
		return err
	}

	total := len(result.Objects)
	if total == 0 {
		v.logger.Warningf("No objects found in the snapshot: %q", result)
		return nil
	}

	v.logger.Infof("Loaded objects: %d", total)
	if err := os.MkdirAll(v.target, 0766); err != nil {
		return err
	}

	siteToLocal := buildMapping(result)
	v.logger.Debugf("Link mapping done, got records: %d", len(siteToLocal))

	for i, obj := range result.Objects {
		v.logger.Infof("[%06d of %06d] object exporting %q", i, total, obj.Id)
		body := strings.Builder{}
		for _, content := range obj.Content {
			body.WriteString(content.Text)
		}
		if _, err := v.exportFile(obj.Id, "text/plain",
			bytes.NewBufferString(convertLinks(body.String(), siteToLocal[obj.Id]))); err != nil {
			v.logger.Errorf("[%06d of %06d] error exporting %q: %q", i, total, obj.Id, err)
		}
	}
	v.logger.Debugf("Exported all %d objects", len(result.Objects))

	for i, obj := range result.Objects {
		mapping := siteToLocal[obj.Id]
		for _, att := range obj.Attachment {
			attUrl := att.Url.Href
			fileTarget := filepath.Join(v.target, common.SanitizeWithExt(attUrl, att.Mime, 255))
			if _, err := os.Stat(fileTarget); err == nil {
				v.logger.Infof(
					"[%06d of %06d] attachment exporting %q to %q (skipped, already exists)",
					i, total, attUrl, fileTarget)
				continue
			}

			v.logger.Infof(
				"[%06d of %06d] attachment exporting %q to %q", i, total, attUrl, fileTarget)
			request := &storage.GetRequest{Url: attUrl}
			fileBytes, err := store.GetBytes(request)
			if err != nil {
				v.logger.Warningf("cannot open file %q for reading: %q", attUrl, err)
				continue
			}

			var reader io.ReadCloser
			if strings.HasPrefix(att.Mime, "text") {
				reader = io.NopCloser(bytes.NewBufferString(convertLinks(string(fileBytes), mapping)))
			} else {
				reader = io.NopCloser(bytes.NewReader(fileBytes))
			}
			if _, err := v.exportFile(attUrl, "text/plain", reader); err != nil {
				v.logger.Errorf("[%06d of %06d] error attachment exporting %q to %q: %q", i, total,
					obj.Id, fileTarget, err)
			}
		}
	}
	return nil
}
