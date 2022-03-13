package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    
    tgpb "telegram/telegram_pb_go"
)

const (
    token="5133968151:AAGEhhmLpF7vp9z29srJQtJNVBTraBEMsqc"
)

type QueryResponse struct {
    response string
}

type IBot interface {
    
  GetMe() (string, error)
}

type Bot struct {
    IBot

    token string    
}

func (b *Bot) GetMe() (string, error) {
    resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}


func main() {
    user := tgpb.User_builder{}.Build()
    bot := &Bot{}
    
    me, err := bot.GetMe()
    if err != nil {
        fmt.Println("Error: %s", err)
        return
    }
    fmt.Println(me)
}
