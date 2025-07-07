package command

import (
	"fmt"

	"chronicler/adapter"
	"chronicler/adapter/fourchan"
	"chronicler/adapter/pikabu"
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/resolver"
)

func Save(s *Settings, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Save requires one argument: <link>, but got none")
	}
	link, err := opb.ParseLink(args[0])
	if err != nil {
		return err
	}
	logger := common.NewLogger("Save")
	httpClient := common.HttpClientBuilder{
		CookieJarPath: s.HttpSettings.CookieJar,
		CacheDirPath:  s.HttpSettings.CachePath,
	}.Build()
	adapters := []adapter.Adapter{}
	if s.Twitter != nil {
		adapters = append(adapters,
			twitter.NewAdapter(twitter.NewClient(httpClient, s.Twitter.Token)))
		logger.Infof("Twitter adapter loaded")
	}
	if s.Reddit != nil {
		adapters = append(adapters,
			reddit.NewAdapter(httpClient, &reddit.RedditAuth{AccessToken: s.Reddit.Token}))
		logger.Infof("Reddit adapter loaded")
	}
	adapters = append(adapters,
		fourchan.NewAdapter(httpClient),
		pikabu.NewAdapter(httpClient),
		web.NewAdapter(httpClient))
	r := resolver.NewResolver(
		s.Storage.Root,
		common.NewHttpDownloader(httpClient),
		adapters,
	)
	r.Start()
	r.Resolve(link)
	r.Wait()
	r.Stop()
	return nil
}
