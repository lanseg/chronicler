package telegram

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "google.golang.org/protobuf/encoding/protojson"
)

type IBot interface {
    
  GetMe() (*User, error)
}

type telegramBot struct {
    IBot

    token string    
}

func queryApi(token string, apiMethod string) ([]byte, error) {
    resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, apiMethod))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    return body, nil
}

func NewBot(token string) IBot {
    return &telegramBot {
        token: token,
    }
}

func (b *telegramBot) GetMe() (*User, error) {
    response, err := queryApi(b.token, "getMe")
    if err != nil {
        return nil, err
    }

    responseProto := &GetMeResponse{}
    err = protojson.Unmarshal(response, responseProto)
    if err != nil {
        return nil, err
    }
    if !responseProto.Ok {
        return nil, fmt.Errorf("Cannot read the result (%d): %s", responseProto.ErrorCode, responseProto.Description)
    }
    
    return responseProto.Result, nil
}
