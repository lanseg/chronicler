package exporter

import (
	"strings"

	"chronicler/common"
	opb "chronicler/proto"
)

type Exporter interface {
	Export(destination string) error
}

func sanitizeURL(maybeUrl string, mimeType string) string {
	return common.SanitizeWithExt(maybeUrl, mimeType, 255)
}

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
				obj.Id: sanitizeURL(obj.Id, ""),
			}
		}
		currentMapping := siteToLocal[obj.Id]

		for _, att := range obj.Attachment {
			safeUrl := sanitizeURL(att.Url.Href, att.Mime)
			for _, variant := range append(att.Url.Variants, att.Url.Href) {
				currentMapping[variant] = safeUrl
			}
		}
	}
	return siteToLocal
}
