package main

import (
	"flag"
	"fmt"
	"strings"

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

	tweets := []*twitter.Tweet{}
	result, err := client.GetConversation("1605769469833494529")
	if err != nil {
	    logger.Errorf("Cannot load tweet: %s", err)
	}
    tweets = append(tweets, result.Data...)
    tweets = append(tweets, result.Includes.Tweets...)
    
	for _, tweet := range tweets {
		fmt.Printf("[%s] %s %s %s\n", tweet.Created, tweet.Id, tweet.ConversationId, strings.ReplaceAll(tweet.Text, "\n", "\\n"))
	}
}
