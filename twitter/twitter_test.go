package twitter

import (
	"chronicler/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

const (
	fakeToken = "fakeToken"
)

type FakeHttpTransport struct {
	http.RoundTripper

	responseIndex int
	responses     []string
}

func (f *FakeHttpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	responseText := f.responses[f.responseIndex]
	f.responseIndex += 1
	return &http.Response{
		Body: io.NopCloser(strings.NewReader(responseText)),
	}, nil
}

func NewFakeClient(responses ...string) Client {
	return &ClientImpl{
		httpClient: &http.Client{
			Transport: &FakeHttpTransport{
				responses: responses,
			},
		},
		token:  fakeToken,
		logger: util.NewLogger("test-twitter"),
	}
}

func readFile(t *testing.T, file string) string {
	result, err := ioutil.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		t.Errorf("Cannot read test data from %s: %s", file, err)
		return ""
	}
	return string(result)
}

func TestTwitterClient(t *testing.T) {

	t.Run("Parse tweet", func(t *testing.T) {
		client := NewFakeClient(readFile(t, "tweet.json"))
		response, err := client.GetTweets([]string{})
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		expected := &Response{
			Data: []*Tweet{
				{
					Id: "1307025659294674945",
					Text: "Here’s an article that highlights the updates in the " +
						"new Tweet payload v2 https://t.co/oeF3ZHeKQQ",
					Created:     "2020-09-18T18:36:15.000Z",
					Author:      "2244994945",
					Attachments: Attachment{},
					Reference: []ReferencedTweet{
						{
							Id:   "1304102743196356610",
							Type: "replied_to",
						},
					},
				},
			},
		}
		if fmt.Sprintf("%s", expected) != fmt.Sprintf("%s", response) {
			t.Errorf("Expected %s, but got %s", expected, response)
		}
	})
}
