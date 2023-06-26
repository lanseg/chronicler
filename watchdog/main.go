package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"chronicler"
	"chronicler/storage"
	"chronicler/telegram"

	//"chronicler/twitter"
	"chronicler/util"

	rpb "chronicler/proto/records"

	"github.com/lanseg/golang-commons/collections"
)

var (
	log = util.NewLogger("main")
)

func parseRequest(s string) rpb.Request {
	source := &rpb.Source{}
	if parsedUrl, err := url.ParseRequestURI(s); err == nil {
		source.Url = s
		source.ChannelId = fmt.Sprintf("%s_%x",
			parsedUrl.Host,
			md5.Sum([]byte(parsedUrl.String())))
		source.Type = rpb.SourceType_WEB
	}

	re := regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
	matches := collections.NewMap(re.SubexpNames(), re.FindStringSubmatch(s))
	if match, ok := matches["twitter_id"]; ok && match != "" {
		source.ChannelId = matches["twitter_id"]
		source.Type = rpb.SourceType_TWITTER
		source.Url = s
	}
	return rpb.Request{Source: source}
}

func fromTelegramUpdate(upd *telegram.Update, baseRecord *rpb.Record) *rpb.Record {
	msg := upd.Message
	result := &rpb.Record{
		Source: &rpb.Source{
			SenderId:  fmt.Sprintf("%d", msg.From.ID),
			ChannelId: fmt.Sprintf("%d", msg.Chat.ID),
			MessageId: fmt.Sprintf("%d", msg.MessageID),
			Type:      rpb.SourceType_TELEGRAM,
		},
		Time: msg.Date,
	}
	for _, e := range msg.Entities {
		if e.Type == "url" && e.URL != "" {
			result.Links = append(result.Links, e.URL)
		}
	}

	result.TextContent = strings.Replace(msg.Caption+"\n"+msg.Text, "\n\n", "\n", -1)
	if msg.Video != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Video.FileID})
	}
	if msg.Photo != nil {
		result.Files = append(result.Files, &rpb.File{
			FileId: telegram.GetLargestImage(msg.Photo).FileID,
		})
	}
	if msg.Audio != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Audio.FileID})
	}
	if msg.Voice != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Voice.FileID})
	}
	if baseRecord != nil {
		result.Links = append(result.Links, baseRecord.Links...)
		result.Files = append(result.Files, baseRecord.Files...)
		newText := result.TextContent
		if strings.Contains(baseRecord.TextContent, newText) {
			newText = baseRecord.TextContent
		} else if !strings.Contains(newText, baseRecord.TextContent) {
			newText += "\n" + baseRecord.TextContent
		}
		result.TextContent = newText
	}
	result.Links = collections.Unique(append(result.Links, util.FindWebLinks(result.TextContent)...))
	sort.Strings(result.Links)
	return result
}

func fetchTelegram(tg *telegram.Bot, cursor int64) ([]*rpb.Record, int64, error) {
	var updates []*telegram.Update = nil
	records := map[string]*rpb.Record{}
	for len(updates) == 0 {
		updates, _ = tg.GetUpdates(int64(0), cursor, 100, 100, []string{})
		log.Debugf("Got %d updates from Telegram", len(updates))
		for _, upd := range updates {
			if cursor < upd.UpdateID {
				cursor = upd.UpdateID
			}
			if upd.Message == nil {
				continue
			}
			msg := upd.Message
			key := fmt.Sprintf("%d_%d_%d", msg.Chat.ID, msg.From.ID, msg.Date)
			records[key] = fromTelegramUpdate(upd, records[key])
		}
	}
	log.Debugf("Done resolving updates.")
	for _, record := range records {
		if len(record.Files) == 0 {
			continue
		}
		for _, file := range record.Files {
			log.Debugf("Loading file for %s", file)
			fileURL, err := tg.GetFile(file.FileId)
			if err != nil {
				log.Errorf("Cannot get actual file url for %s: %s", file.FileId, err)
				continue
			}
			file.FileUrl = tg.GetUrl(fileURL)
		}
	}
	return collections.Values(records), cursor, nil
}

func main() {
	flag.Parse()
	cfg := chronicler.GetConfig()
	stg := storage.NewStorage(*cfg.StorageRoot)
	chr := chronicler.NewWeb("web", nil)

	tgbot := telegram.NewBot(*cfg.TelegramBotKey)
	cursor := int64(0)
	for {
		log.Infof("Waiting for the next request, cursor: %d", cursor)
		newRequests, nextCursor, err := fetchTelegram(tgbot, cursor)
		if err != nil {
			continue
		}
		cursor = nextCursor + 1
		for _, request := range newRequests {
			log.Infof("Got request: %s", request)
			result := &rpb.RecordSet{
				Request: &rpb.Request{
					Source: request.Source,
				},
				Records: []*rpb.Record{request},
			}
			for _, arg := range request.Links {
				if arg == "" {
					continue
				}
				req := parseRequest(arg)
				log.Infof("Loading attached request: %s", req)
				conv, err := chr.GetRecords(&req)
				if err != nil {
					log.Errorf("Failed to get conversation for id %s: %s", request, err)
					continue
				}
				conv.Request = &req
				result.Records = append(result.Records, conv.Records...)
			}
			replyToId, _ := strconv.Atoi(result.Request.Source.SenderId)
			if err := stg.SaveRecords(result); err != nil {
				log.Warningf("Error while saving a record: %s", err)
				tgbot.SendMessage(int64(replyToId), 0, fmt.Sprintf("Failed saving records: %s", err))
			} else {
				tgbot.SendMessage(int64(replyToId), 0, fmt.Sprintf("Records saved: %d", len(result.Records)))
			}
		}
	}
}
