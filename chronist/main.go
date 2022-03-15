package main

import (
    "fmt"
    "os"
    "chronist/telegram"
)

func main() {
    b := telegram.NewBot(os.Args[1])
    updates, err := b.GetUpdates(0, 10, 1, []string{"message"})
    if err != nil {
      fmt.Println(err)
    } else {
        for _, upd := range updates {
          fmt.Println(upd)
        }
    }
}
