package telegram

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/tgbot"

	rpb "chronicler/records/proto"
)

const (
	testingUuid = "1a468cef-1368-408a-a20b-86b32d94a460"
)

func writeJson(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func readJson[T any](file string) (*T, error) {
	bytes, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return nil, err
	}
	return cm.FromJson[T](bytes)
}

type fakeBot struct {
	tgbot.TelegramBot

	responded bool
	updates   string
}

func (b *fakeBot) Query(methodName string, body interface{}) ([]byte, error) {
	switch methodName {
	case "GetUpdates":
		return os.ReadFile(filepath.Join("testdata", b.updates))
	case "SendMessage":
		return []byte("{\"ok\": true}"), nil
	case "GetFile":
		return []byte("{\"ok\": true, \"result\": { \"file_id\": \"file_id_here\", \"file_unique_id\": \"file_unique_id_here\", \"file_path\": \"file_path_here\" }}"), nil
	}
	return []byte("{\"ok\": true}"), nil
}

func (b *fakeBot) ResolveUrl(path string) string {
	return fmt.Sprintf("https://telegram/url/%s", path)
}

func NewFakeBot(datafile string) tgbot.TelegramBot {
	return &fakeBot{updates: datafile}
}

func TestRequestResponse(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		responseFile string
		resultFile   string
	}{
		{
			desc:         "Single update response",
			responseFile: "telegram_one_update.json",
			resultFile:   "telegram_one_update_record.json",
		},
		{
			desc:         "Multiple update response",
			responseFile: "telegram_multiple_updates.json",
			resultFile:   "telegram_multiple_updates_record.json",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tg := NewTelegramAdapter(NewFakeBot(tc.responseFile))
			ups := tg.GetResponse(&rpb.Request{Id: testingUuid})[0].Result[0]
			ups.Id = testingUuid
			for _, r := range ups.Records {
				r.FetchTime = 0
			}

			os.WriteFile("/tmp/"+tc.resultFile+"_result", []byte(writeJson(ups)), 0644)
			want, err := readJson[rpb.RecordSet](tc.resultFile)
			if err != nil {
				t.Errorf("Cannot load json with an expected result \"%s\": %s", tc.resultFile, err)
			}
			if fmt.Sprintf("%+v", want) != fmt.Sprintf("%+v", ups) {
				t.Errorf("Expected result to be:\n%s\nBut got:\n%s", writeJson(want), writeJson(ups))
			}
		})
	}
}
