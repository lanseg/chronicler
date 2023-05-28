package main

import (
	"fmt"
	"os"

	"chronicler/twitter"
	"encoding/json"
)

func pretty(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func main() {
	fmt.Println("Twitter experimental")

	client := twitter.NewClient(os.Args[1])
	tweets, err := client.GetTweets([]string{""})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	str, err := pretty(tweets)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("Result: %s\n", str)
	}
}
