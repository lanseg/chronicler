package reddit

import (
	"fmt"
	"testing"
)

func TestRedditClient(t *testing.T) {
	client := &redditClient{}
	fmt.Printf("HERE: %s\n", client.GetPost(nil))
}
