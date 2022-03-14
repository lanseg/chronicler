package main

import (
    "fmt"
    "os"
    "chronist/telegram"
)

func main() {
    b := telegram.NewBot(os.Args[1])
    me, err := b.GetMe()
    if err != nil {
        fmt.Println(err)
    } else {
      fmt.Println(me)
    }
}
