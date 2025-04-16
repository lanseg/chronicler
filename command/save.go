package command

import (
	"chronicler/adapter"
	"chronicler/adapter/fourchan"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/resolver"
	"chronicler/storage"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func Save(tw *twitter.Settings, rs *reddit.Settings, st *storage.Settings, args []string) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Minute,
	}

	r := resolver.NewResolver(
		st.Root,
		common.NewHttpDownloader(httpClient),
		[]adapter.Adapter{
			twitter.NewAdapter(twitter.NewClient(httpClient, tw.Token)),
			fourchan.NewAdapter(httpClient),
			pikabu.NewAdapter(httpClient),
			reddit.NewAdapter(httpClient, &reddit.RedditAuth{AccessToken: rs.Token}),
			web.NewAdapter(httpClient),
		},
	)
	r.Start()
	r.Resolve(&opb.Link{Href: args[0]})
	r.Wait()
	r.Stop()
}
