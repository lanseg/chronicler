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
	httpClient := common.HttpClientBuilder{
		CookieJarPath: s.HttpSettings.CookieJar,
		CacheDirPath:  s.HttpSettings.CachePath,
	}.Build()

	return resolver.NewResolver(
		s.Storage.Root,
		common.NewHttpDownloader(httpClient),
		[]adapter.Adapter{
			twitter.NewAdapter(httpClient, s.Twitter),
			reddit.NewAdapter(httpClient, s.Reddit),
			fourchan.NewAdapter(httpClient),
			pikabu.NewAdapter(httpClient),
			web.NewAdapter(httpClient, s.WebSettings),
		},
	).Resolve(link)
}
