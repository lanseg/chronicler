package pikabu

import (
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"chronicler/common"
	"chronicler/parser"

	opb "chronicler/proto"
)

const (
	InError          = -1
	InDocument       = 0
	InArticle        = 1
	InArticleTitle   = 2
	InArticleContent = 3
	InArticleTags    = 4
	InArticleRating  = 5
	InComment        = 6
	InCommentHeader  = 7
	InCommentContent = 8
	InCommentRating  = 9
	InAfterComments  = 10
)

func toTimestamp(s string) (*opb.Timestamp, error) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return &opb.Timestamp{Seconds: int64(i)}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &opb.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}, nil
	}
	return nil, err
}

type PikabuParser struct {
	doc    parser.HtmlReader
	logger *common.Logger

	count  int
	state  int
	states map[int]func()

	article    *opb.Object
	comment    []*opb.Object
	attachment map[string]*opb.Attachment
}

func (psm *PikabuParser) newArticle() {
	psm.article = &opb.Object{
		Id: psm.doc.Attr("data-story-id"),
		Generator: []*opb.Generator{
			{
				Id:   psm.doc.Attr("data-author-id"),
				Name: psm.doc.Attr("data-author-name"),
			},
		},
	}
	if len(psm.attachment) != 0 || psm.attachment == nil {
		psm.attachment = map[string]*opb.Attachment{}
	}
}

func (psm *PikabuParser) newComment() {
	meta := map[string]string{}
	for _, metaParam := range strings.Split(psm.doc.Attr("data-meta"), ";") {
		kv := strings.Split(metaParam, "=")
		if len(kv) == 1 {
			meta[kv[0]] = ""
		} else {
			meta[kv[0]] = kv[1]
		}
	}
	if meta["pid"] == "0" && psm.article != nil {
		meta["pid"] = psm.article.Id
	}
	t, err := toTimestamp(meta["d"])
	if err != nil {
		// TODO: log time parse error
	}
	psm.comment = append(psm.comment, &opb.Object{
		Id:        psm.doc.Attr("data-id"),
		CreatedAt: t,
		Parent:    meta["pid"],
		Generator: []*opb.Generator{
			{
				Id: psm.doc.Attr("data-author-id"),
			},
		},
	})
	if len(psm.attachment) != 0 || psm.attachment == nil {
		psm.attachment = map[string]*opb.Attachment{}
	}
}

func (psm *PikabuParser) getAttachments() {
	urls := map[string]bool{}
	for _, attr := range []string{
		"href", "src", "data-src", "data-source", "data-hls", "data-webm", "data-mp4",
		"data-large-image", "data-url", "data-thumb",
	} {
		if attrValue := psm.doc.Attr(attr); attrValue != "" {
			maybeUrl, err := url.Parse(attrValue)
			if err != nil {
				psm.logger.Warningf("Cannot parse %q as url: %s", attrValue, err)
				continue
			}
			if attr == "href" && maybeUrl.Query().Has("u") {
				actualUrl := maybeUrl.Query().Get("u")
				attrValue, err = url.QueryUnescape(actualUrl)
				if err != nil {
					psm.logger.Warningf("Cannot decode %q as url: %s", actualUrl, err)
					continue
				}
			}
			urls[attrValue] = true
		}
	}
	for k := range urls {
		attachment := &opb.Attachment{Url: k}
		if ext := filepath.Ext(k); ext != "" {
			attachment.Mime = mime.TypeByExtension(filepath.Ext(k))
		}
		psm.attachment[k] = attachment
	}
}

func (psm *PikabuParser) InDocument() {
	if psm.doc.Matches("article") {
		if psm.article == nil {
			psm.newArticle()
			psm.SetState(InArticle)
		}
	} else if psm.doc.Matches("div", "comment") {
		psm.newComment()
		psm.SetState(InComment)
	}
}

func (psm *PikabuParser) InArticle() {
	if psm.doc.Matches("span", "story__title-link") {
		psm.SetState(InArticleTitle)
	} else if psm.doc.Matches("div", "story__tags") {
		psm.SetState(InArticleTags)
	} else if psm.doc.Matches("div", "story__content-inner") {
		psm.count = 1
		psm.article.Content = append(psm.article.Content, &opb.Content{
			Mime: "text/html",
		})
		psm.SetState(InArticleContent)
	} else if psm.doc.Matches("div", "story__rating-count") {
		psm.SetState(InArticleRating)
	} else if psm.doc.Matches("time") {
		if ts, err := toTimestamp(psm.doc.Attr("datetime")); err == nil {
			psm.article.CreatedAt = ts
		}
	} else if psm.doc.Matches("/article") {
		psm.SetState(InDocument)
	}
}

func (psm *PikabuParser) InArticleTitle() {
	if psm.doc.Matches("/span") {
		psm.SetState(InArticle)
	} else {
		psm.article.Content = append(psm.article.Content, &opb.Content{Text: psm.doc.Raw()})
	}
}

func (psm *PikabuParser) InArticleRating() {
	if psm.doc.Matches("/div") {
		psm.SetState(InArticle)
	} else {
		rating, err := strconv.ParseInt(psm.doc.Raw(), 10, 64)
		if err != nil {
			return
		}
		psm.article.Stats = append(psm.article.Stats, &opb.Stats{
			Type:    opb.Stats_RATING,
			Counter: rating,
		})
	}
}

func (psm *PikabuParser) InArticleContent() {
	if psm.doc.Matches("div") {
		psm.count++
	} else if psm.doc.Matches("/div") {
		psm.count--
	}
	if psm.count == 0 {
		for _, attachment := range psm.attachment {
			psm.article.Attachment = append(psm.article.Attachment, attachment)
		}
		psm.SetState(InArticle)
		return
	}
	psm.article.Content[len(psm.article.Content)-1].Text += psm.doc.Raw()
	psm.getAttachments()
}

func (psm *PikabuParser) InArticleTags() {
	if psm.doc.Matches("/div") {
		psm.SetState(InArticle)
	} else if psm.doc.Matches("a", "tags__tag") {
		psm.article.Tag = append(psm.article.Tag, &opb.Tag{
			Name: psm.doc.Attr("data-tag"),
			Url:  psm.doc.Attr("href"),
		})
	}
}

func (psm *PikabuParser) InComment() {
	if psm.doc.Matches("div", "comment__header") {
		psm.SetState(InCommentHeader)
		psm.count = 1
	} else if psm.doc.Matches("div", "comment__content") {
		psm.count = 1
		psm.SetState(InCommentContent)
	}
}

func (psm *PikabuParser) InCommentHeader() {
	if psm.doc.HasClass("comment__rating-count") {
		psm.SetState(InCommentRating)
		return
	}
	if psm.doc.Matches("div") {
		psm.count++
	} else if psm.doc.Matches("/div") {
		psm.count--
	}
	if psm.count == 0 {
		psm.SetState(InComment)
		return
	}
	lastComment := psm.comment[len(psm.comment)-1]
	if psm.doc.Matches("div", "comment__user") {
		lastComment.Generator[0].Name = psm.doc.Attr("data-name")
	}
}

func (psm *PikabuParser) InCommentRating() {
	if psm.doc.Matches("/div") {
		psm.SetState(InCommentHeader)
	} else {
		rating, err := strconv.ParseInt(psm.doc.Raw(), 10, 64)
		if err != nil {
			return
		}
		lastComment := psm.comment[len(psm.comment)-1]
		lastComment.Stats = append(lastComment.Stats, &opb.Stats{
			Type:    opb.Stats_RATING,
			Counter: rating,
		})
	}
}

func (psm *PikabuParser) InCommentContent() {
	lastComment := psm.comment[len(psm.comment)-1]
	if psm.doc.Matches("div") {
		psm.count++
	} else if psm.doc.Matches("/div") {
		psm.count--
	}
	if psm.count == 0 {
		psm.SetState(InDocument)
		for _, attachment := range psm.attachment {
			lastComment.Attachment = append(lastComment.Attachment, attachment)
		}
		return
	}
	if lastComment.Content == nil {
		lastComment.Content = []*opb.Content{{Mime: "text/html"}}
	}
	psm.getAttachments()
	lastComment.Content[len(lastComment.Content)-1].Text += string(psm.doc.Raw())
}

func (psm *PikabuParser) SetState(state int) error {
	psm.state = state
	return nil
}

func (psm *PikabuParser) Parse() ([]*opb.Object, error) {
	for !psm.doc.NextToken() {
		psm.states[psm.state]()
	}
	result := []*opb.Object{}
	if psm.article != nil {
		result = append(result, psm.article)
	}
	result = append(result, psm.comment...)
	return result, nil
}

func NewPikabuParser(src io.Reader) *PikabuParser {
	psm := &PikabuParser{
		doc:     parser.NewHtmlReader(src),
		state:   InDocument,
		comment: []*opb.Object{},
		logger:  common.NewLogger("PikabuParser"),
	}
	psm.states = map[int]func(){
		InDocument:       psm.InDocument,
		InArticle:        psm.InArticle,
		InArticleTitle:   psm.InArticleTitle,
		InArticleContent: psm.InArticleContent,
		InArticleTags:    psm.InArticleTags,
		InArticleRating:  psm.InArticleRating,
		InComment:        psm.InComment,
		InCommentRating:  psm.InCommentRating,
		InCommentHeader:  psm.InCommentHeader,
		InCommentContent: psm.InCommentContent,
	}
	return psm
}
