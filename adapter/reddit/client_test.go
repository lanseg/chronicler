package reddit

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
)

func TestRedditClient(t *testing.T) {
	ad := NewAdapter(adaptertest.NewFakeHttp("/home/arusakov/devel/lanseg/chronicler/reddit.json"))

	es, err := ad.Get(&opb.Link{Href: "https://www.reddit.com/r/AnimalsBeingDerps/comments/1ioa2mb"})
	if err != nil {
		fmt.Printf("HERE Error while loading reddit post: %s", err)
	} else {
		bytes, err := json.Marshal(es)
		if err != nil {
			fmt.Printf("HERE %s\n", err)
			return
		}
		fmt.Printf("HERE: %s\n",
			os.WriteFile("/home/arusakov/devel/lanseg/chronicler/reddit-out.json", bytes, 0777))
	}
}
