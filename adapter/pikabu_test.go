package adapter

import (
	"fmt"
	"os"
	"testing"

	rpb "chronicler/records/proto"
	"chronicler/webdriver"
)

const (
	pikabuWithComments = "pikabu_10819340.html"
)

func TestPikabuRequestResponse(t *testing.T) {

	t.Run("Creates", func(t *testing.T) {
		pkb := NewPikabuAdapter(webdriver.WrapExclusive(newFakeWebdriver(pikabuWithComments)))
		resp := pkb.GetResponse(&rpb.Request{
			Target: pkb.MatchLink("https://pikabu.ru/story/pikabu_some_123123.html"),
		})
		fmt.Printf("HERE: %s\n", os.WriteFile("/tmp/pkb.json", []byte(writeJson(resp)), os.ModePerm))
	})
}
