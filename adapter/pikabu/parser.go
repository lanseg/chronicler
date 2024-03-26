package pikabu

import (
	"fmt"
	"strings"
	"time"

	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/almosthtml"
	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

func parseStory(node *almosthtml.Node) (*rpb.Record, *rpb.UserMetadata) {
	author := &rpb.UserMetadata{}
	result := &rpb.Record{
		FetchTime: time.Now().Unix(),
		Source: &rpb.Source{
			Type: rpb.SourceType_PIKABU,
		},
	}
	inContent := false
	textContent := strings.Builder{}
	collections.IterateTree(node, collections.DepthFirst, func(n *almosthtml.Node) []*almosthtml.Node {
		class := n.Params["class"]
		if n.Name == "div" && (class == "story__tags tags" || class == "story__footer") {
			return []*almosthtml.Node{}
		}
		return n.Children
	}).ForEachRemaining(func(n *almosthtml.Node) bool {
		class := n.Params["class"]
		dataName, hasDataName := n.Params["data-name"]
		if class == "story__user-link user__nick" && hasDataName {
			author.Username = dataName
			author.Id = n.Params["data-id"]
			result.Source.SenderId = author.Id
		} else if class == "image-link" {
			result.Files = append(result.Files, &rpb.File{
				FileId:  cm.UUID4For(n.Params["href"]),
				FileUrl: n.Params["href"],
			})
		} else if n.Name == "a" && n.Params["href"] != "" {
			result.Links = append(result.Links, n.Params["href"])
		}
		if dataLargeImage, hasLargeImage := n.Params["data-large-image"]; hasLargeImage {
			result.Files = append(result.Files, &rpb.File{
				FileId:  cm.UUID4For(dataLargeImage),
				FileUrl: dataLargeImage,
			})
		}
		if src, hasSrc := n.Params["src"]; hasSrc && n.Name == "source" {
			result.Files = append(result.Files, &rpb.File{
				FileId:  cm.UUID4For(src),
				FileUrl: src,
			})
		}

		if n.Name == "time" {
			t, err := time.Parse(time.RFC3339, strings.Replace(n.Params["datetime"], "t", "T", -1))
			if err == nil {
				result.Time = t.Unix()
			}
			return false
		}
		if class == "story__content story__typography" {
			inContent = true
		}
		if class == "tags__tag" {
			inContent = false
		}
		if !inContent {
			return false
		}
		if n.Name == "br" || n.Name == "p" {
			textContent.WriteRune('\n')
		}
		if n.Name == "#text" {
			text := strings.ReplaceAll(strings.TrimSpace(n.Raw), "\n", "")
			if text != "" {
				textContent.WriteString(text)
			}
		}
		return false
	})
	result.Source = &rpb.Source{
		SenderId: author.Id,
		Type:     rpb.SourceType_PIKABU,
	}
	result.TextContent = strings.TrimSpace(textContent.String())
	return result, author
}

func parseCommentContent(node *almosthtml.Node) (string, []string, []string) {
	result := strings.Builder{}
	links := []string{}
	files := []string{}
	collections.IterateTree(node, collections.DepthFirst, func(n *almosthtml.Node) []*almosthtml.Node {
		return n.Children
	}).ForEachRemaining(func(n *almosthtml.Node) bool {
		if href, hasHref := n.Params["href"]; hasHref {
			links = append(links, href)
		}
		if thumb, hasThumb := n.Params["data-thumb"]; hasThumb {
			files = append(files, thumb)
		}
		if class := n.Params["class"]; n.Name == "a" && class == "image-link" {
			files = append(files, n.Params["href"])
		}
		if url, hasUrl := n.Params["data-url"]; hasUrl {
			links = append(links, url)
		}
		if n.Name == "br" || n.Name == "p" {
			result.WriteRune('\n')
		}
		if n.Name == "#text" {
			text := strings.ReplaceAll(strings.TrimSpace(n.Raw), "\n", "")
			if text != "" {
				result.WriteString(text)
			}
		}
		return false
	})
	return strings.TrimSpace(result.String()), links, files
}

func parseComment(n *almosthtml.Node) (*rpb.Record, *rpb.UserMetadata) {
	meta := map[string]string{}
	for _, m := range strings.Split(n.Params["data-meta"], ";") {
		params := strings.Split(m, "=")
		if len(params) == 2 {
			meta[params[0]] = params[1]
		} else {
			meta[params[0]] = ""
		}
	}

	result := &rpb.Record{
		FetchTime: time.Now().Unix(),
		Source: &rpb.Source{
			SenderId:  meta["aid"],
			ChannelId: meta["sid"],
			MessageId: n.Params["data-id"],
			Type:      rpb.SourceType_PIKABU,
		},
	}

	if meta["pid"] != "0" {
		result.Parent = &rpb.Source{
			MessageId: meta["pid"],
			ChannelId: meta["sid"],
			Type:      rpb.SourceType_PIKABU,
		}
	}

	bodies := n.GetElementsByTagAndClass("div", "comment__body")
	if len(bodies) == 0 {
		return nil, nil
	}
	body := n.GetElementsByTagAndClass("div", "comment__body")[0]
	header := body.GetElementsByTagAndClass("div", "comment__header")[0]
	user := header.GetElementsByTagAndClass("div", "comment__user")[0]
	tvalue := header.GetElementsByTagAndClass("time")[0]
	t, err := time.Parse(time.RFC3339, strings.Replace(tvalue.Params["datetime"], "t", "T", -1))
	if err == nil {
		result.Time = t.Unix()
	}

	content, links, files := parseCommentContent(body.GetElementsByTagAndClass("div", "comment__content")[0])

	result.Links = append(result.Links, links...)
	for _, f := range files {
		result.Files = append(result.Files, &rpb.File{
			FileId:  cm.UUID4For(f),
			FileUrl: f,
		})
	}
	result.TextContent = content
	userData := &rpb.UserMetadata{
		Id:       user.Params["data-id"],
		Username: user.Params["data-name"],
	}

	return result, userData
}

func parsePost(content string) (*rpb.Response, error) {
	resultRecords := []*rpb.Record{}
	userById := map[string]*rpb.UserMetadata{}
	commentById := map[string]*rpb.Source{}
	root, _ := almosthtml.ParseHTML(content)
	if n := root.GetElementsByTags("title"); len(n) != 0 {
		title := n[0].InnerHTML()
		if strings.Contains(title, "Страница удалена") || strings.Contains(title, "Страница не найдена") {
			return nil, fmt.Errorf("Page was removed: %s", title)
		}
	}
	for _, n := range root.GetElementsByTagAndClass("div") {
		if n.Params["class"] == "story__main" {
			story, author := parseStory(n)
			resultRecords = append(resultRecords, story)
			userById[author.Id] = author
		}
		if n.Params["class"] == "section-hr" {
			break
		}
		if n.Params["data-meta"] != "" && n.Params["class"] == "comment" {
			comment, user := parseComment(n)
			if comment == nil {
				continue
			}
			if comment.Parent == nil {
				comment.Parent = resultRecords[0].Source
			}
			resultRecords = append(resultRecords, comment)
			commentById[comment.Source.MessageId] = comment.Source
			userById[user.Id] = user
		}
	}
	for _, r := range resultRecords {
		if r.Parent == nil {
			continue
		}
		if parent := commentById[r.Parent.MessageId]; parent != nil {
			r.Parent = parent
		}
	}

	return &rpb.Response{
		Result: []*rpb.RecordSet{
			{
				Records:      resultRecords,
				UserMetadata: collections.Values(userById),
			},
		},
	}, nil
}
