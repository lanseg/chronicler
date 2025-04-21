package command

import (
	"bytes"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func denormalize(root *url.URL, link string) []string {
	result := []string{link}
	url, err := url.Parse(link)
	if err != nil {
		return result
	}
	if url.Scheme != "" {
		url.Scheme = ""
		result = append(result, url.String())
	}
	if root != nil && url.Host == root.Host {
		url.Host = ""
		pathOnly := url.String()
		if pathOnly != "/" && pathOnly != "" {
			result = append(result, url.String())
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return len(result[j]) < len(result[i])
	})
	return result
}

func convertLinks(text string, realToLocal map[string][]string) string {
	for localPath, links := range realToLocal {
		for _, link := range links {
			text = strings.ReplaceAll(text, fmt.Sprintf("\"%s\"", link), fmt.Sprintf("\"%s\"", localPath))
			text = strings.ReplaceAll(text, fmt.Sprintf("'%s'", link), fmt.Sprintf("'%s'", localPath))
		}
	}
	return text
}

type Exporter struct {
	Root   string
	Target string
	logger *common.Logger
}

func NewExporter(root string, target string) *Exporter {
	return &Exporter{
		Root:   root,
		Target: target,
		logger: common.NewLogger("export"),
	}
}

func (v *Exporter) Export(id string) error {
	store := storage.BlockStorage{
		Storage: common.OrExit(storage.NewLocalStorage(filepath.Join(v.Root, id))),
	}
	v.logger.Infof("Loading objects from %q", filepath.Join(v.Root, id, objectFileName))
	result := &opb.Snapshot{}
	if err := store.GetObject(&storage.GetRequest{Url: objectFileName}, &result); err != nil {
		return err
	}
	root, _ := url.Parse(result.Link.Href)
	total := len(result.Objects)
	v.logger.Infof("Loaded objects: %d", total)
	os.MkdirAll(v.Target, 0766)
	for i, obj := range result.Objects {
		if !strings.Contains(obj.Id, "scp-series") {
			continue
		}
		atmapping := map[string][]string{}
		for _, att := range obj.Attachment {
			safeUrl := common.SanitizeUrl(att.Url, 255)
			exts, err := mime.ExtensionsByType(att.Mime)
			if err == nil && len(exts) > 0 {
				safeUrl += exts[0]
			}
			atmapping[safeUrl] = denormalize(root, att.Url)
			fmt.Printf("HERE %q -> %q\n", safeUrl, atmapping[safeUrl])
		}

		fileTarget := filepath.Join(v.Target, common.SanitizeUrl(obj.Id, 255)+".html")
		v.logger.Infof("[%06d of %06d] exporting %q to %q", i, total, obj.Id, fileTarget)
		f, err := os.OpenFile(fileTarget, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			v.logger.Warningf("cannot open file %q for writing: %q", fileTarget, err)
			continue
		}
		for _, content := range obj.Content {
			text := convertLinks(content.Text, atmapping)
			if _, err := f.Write([]byte(text)); err != nil {
				v.logger.Warningf("cannot write to file %q: %q", fileTarget, err)
				break
			}
		}
		f.Close()

		for _, att := range obj.Attachment {
			safeUrl := common.SanitizeUrl(att.Url, 255)
			exts, err := mime.ExtensionsByType(att.Mime)
			if err == nil && len(exts) > 0 {
				safeUrl += exts[0]
			}
			fileTarget := filepath.Join(v.Target, safeUrl)
			if _, err := os.Stat(fileTarget); err == nil {
				v.logger.Infof("[%06d of %06d] exporting %q to %q (skipped, already exists)", i, total, att.Url, fileTarget)
				continue
			}
			v.logger.Infof("[%06d of %06d] exporting %q to %q", i, total, att.Url, fileTarget)
			request := &storage.GetRequest{Url: att.Url}
			var reader io.ReadCloser
			if strings.HasPrefix(att.Mime, "text") {
				fileBytes, err := store.GetBytes(request)
				if err != nil {
					v.logger.Warningf("cannot open file %q for reading: %q", att.Url, err)
					continue
				}
				reader = io.NopCloser(bytes.NewReader([]byte(convertLinks(string(fileBytes), atmapping))))
			} else {
				fileBytes, err := store.GetBytes(request)
				if err != nil {
					v.logger.Warningf("cannot open file %q for reading: %q", att.Url, err)
					continue
				}
				reader = io.NopCloser(bytes.NewReader([]byte(convertLinks(string(fileBytes), atmapping))))
			}

			f, err := os.Create(fileTarget)
			if err != nil {
				v.logger.Warningf("cannot open file %q for writing: %q", fileTarget, err)
				continue
			}
			if _, err := io.Copy(f, reader); err != nil {
				v.logger.Warningf("cannot write %q to %q: %q", att.Url, fileTarget, err)
			}
			f.Close()
			reader.Close()
		}
		break
	}
	return nil
}
