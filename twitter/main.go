package main

import (
	"flag"
	"fmt"

	"chronist/twitter"
	"chronist/util"
)

const (
	twitterApiFlag = "twitter_api_key"
)

var (
	twitterApiKey = flag.String(twitterApiFlag, "", "A key for the twitter api.")
)

func main() {
	flag.Parse()
	logger := util.NewLogger("main")

	client := twitter.NewClient(*twitterApiKey)
	token := ""
	for {
		result, err := client.GetConversation("1605769469833494529", token)
		if err != nil {
			logger.Errorf("Cannot load tweet: %s", err)
			return
		}
		for _, tweet := range result.Data {
			fmt.Printf("%s %s %s\n", tweet.Id, tweet.ConversationId, tweet.Text)
		}
		token = result.Meta.NextToken
		if len(result.Data) == 0 || token == "" {
			break
		}
	}
}
