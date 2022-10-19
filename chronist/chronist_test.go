package chronist

import (
	"reflect"
	"testing"

	"chronist/telegram"

	rpb "chronist/proto/records"
)

func TestFromTelegramUpdate(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		update     *telegram.Update
		baseRecord *rpb.Record
		want       *rpb.Record
	}{
		{
			desc: "links from text are also put into the link fields",
			update: &telegram.Update{
				UpdateID: 1234,
				Message: &telegram.Message{
					From: &telegram.User{
						ID: 1234,
					},
					Chat: &telegram.Chat{
						ID: 1234,
					},
					MessageID: 1234,
					Entities: []*telegram.MessageEntity{
						{
							Type: "url",
							URL:  "http://some/url",
						},
					},
					Text: "Hello there https://some.link.text",
				},
			},
			want: &rpb.Record{
				RecordId: "1234",
				Source: &rpb.Source{
					SenderId:  "1234",
					ChannelId: "1234",
					MessageId: "1234",
				},
				TextContent: "Hello there https://some.link.text",
				Links: []string{
					"http://some/url",
					"https://some.link.text",
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := FromTelegramUpdate(tc.update, tc.baseRecord)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("FromTelegramUpdate(%v, %v) expected to be %v, but got %v",
					tc.update, tc.baseRecord, tc.want, result)
			}
		})
	}
}
