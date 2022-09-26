package main

import (
    "fmt"
    "os"
    "chronist/telegram"
    "encoding/json"
)

func main() {
    b := telegram.NewBot(os.Args[1])
    
    updates, _ := b.GetUpdates(986286228, 0, 10, []string{"message", "channel_post", "edited_message", "edited_channel_post"})
    js, _ := json.MarshalIndent(updates, "", "  ")
    fmt.Printf("%s\n", js)

    for _, upd := range updates {
      if upd.Message != nil {
        fmt.Printf("Message: %10s %10s %10s\n", upd.Message.Chat.Title, upd.Message.Chat.Type, upd.Message.Text)
      } else if (upd.ChannelPost != nil) {
        fmt.Printf("ChannelPost: %10s %10s\n", upd.ChannelPost.Chat.Title, upd.ChannelPost.Chat.Type)
      } else {
        fmt.Printf("Other update: %10s\n", "...")
      }
    }
}
