package command

import (
	"bytes"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

func convertLinks(text string, realToLocal map[string]string) string {
	for link, localPath := range realToLocal {
		text = strings.ReplaceAll(strings.ReplaceAll(text, "\""+link+"\"", "\""+localPath+"\""), "'"+link+"'", "'"+localPath+"'")
	}
	return text
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
	v.logger.Infof("Loaded objects: %d", total)
	os.MkdirAll(v.target, 0766)

	atmapping := map[string]map[string]string{}
	for _, obj := range result.Objects {
		if atmapping[obj.Id] == nil {
			atmapping[obj.Id] = map[string]string{}
		}
		currentMapping := atmapping[obj.Id]
		currentMapping[obj.Id] = common.SanitizeUrl(obj.Id, 255)
		for _, att := range obj.Attachment {
			safeUrl := common.SanitizeUrl(att.Url.Href, 255)
			if filepath.Ext(safeUrl) == "" {
				exts, err := mime.ExtensionsByType(att.Mime)
				if err == nil && len(exts) > 0 {
					safeUrl += exts[0]
				}
			}
			currentMapping[att.Url.Href] = safeUrl
			for _, v := range att.Url.Variants {
				currentMapping[v] = safeUrl
			}
		}
	}
	v.logger.Debugf("Link mapping done, records: %d", len(atmapping))

	for i, obj := range result.Objects {
		fileTarget := filepath.Join(v.target, common.SanitizeUrl(obj.Id, 255))
		v.logger.Infof("[%06d of %06d] exporting %q to %q", i, total, obj.Id, fileTarget)
		f, err := os.OpenFile(fileTarget, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			v.logger.Warningf("cannot open file %q for writing: %q", fileTarget, err)
			continue
		}
		mapping := atmapping[obj.Id]
		v.logger.Debugf("Replacing links in %q, to replace: %d", obj.Id, len(mapping))
		for _, content := range obj.Content {
			text := convertLinks(content.Text, mapping)
			if _, err := f.Write([]byte(text)); err != nil {
				v.logger.Warningf("cannot write to file %q: %q", fileTarget, err)
				break
			}
		}
		f.Close()
	}
	v.logger.Debugf("Exported all %d objects", len(result.Objects))

	for i, obj := range result.Objects {
		mapping := atmapping[obj.Id]
		for _, att := range obj.Attachment {
			safeUrl := common.SanitizeUrl(att.Url.Href, 255)
			if filepath.Ext(safeUrl) == "" {
				exts, err := mime.ExtensionsByType(att.Mime)
				if err == nil && len(exts) > 0 {
					safeUrl += exts[0]
				} else {
					safeUrl += ".html"
				}
			}
			fileTarget := filepath.Join(v.target, safeUrl)
			if _, err := os.Stat(fileTarget); err == nil {
				v.logger.Infof("[%06d of %06d] exporting %q to %q (skipped, already exists)", i, total, att.Url.Href, fileTarget)
				continue
			}
			v.logger.Infof("[%06d of %06d] exporting %q to %q", i, total, att.Url.Href, fileTarget)
			request := &storage.GetRequest{Url: att.Url.Href}
			var reader io.ReadCloser
			if strings.HasPrefix(att.Mime, "text") {
				fileBytes, err := store.GetBytes(request)
				if err != nil {
					v.logger.Warningf("cannot open file %q for reading: %q", att.Url.Href, err)
					continue
				}
				reader = io.NopCloser(bytes.NewReader([]byte(convertLinks(string(fileBytes), mapping))))
			} else {
				fileBytes, err := store.GetBytes(request)
				if err != nil {
					v.logger.Warningf("cannot open file %q for reading: %q", att.Url.Href, err)
					continue
				}
				reader = io.NopCloser(bytes.NewReader([]byte(convertLinks(string(fileBytes), mapping))))
			}

			f, err := os.Create(fileTarget)
			if err != nil {
				v.logger.Warningf("cannot open file %q for writing: %q", fileTarget, err)
				continue
			}
			if _, err := io.Copy(f, reader); err != nil {
				v.logger.Warningf("cannot write %q to %q: %q", att.Url.Href, fileTarget, err)
			}
			f.Close()
			reader.Close()
		}
		break
	}
	return nil
}
