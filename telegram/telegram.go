package telegram

import (
    "fmt"
    "strconv"
    "strings"
    "io/ioutil"
    "net/http"
    "net/url"
    "encoding/json"
)

type ResponseMetadata struct {
  Ok bool             `json:"ok"`
  ErrorCode int64     `json:"error_code"`
  Description string  `json:"description"`
}

type GetMeResponse struct {
  ResponseMetadata

  Result User         `json:"result"` 
}

type GetUpdateResponse struct {
  ResponseMetadata

  Result []*Update     `json:"result"` 
}

type IBot interface {
    
  GetMe() (*User, error)
  GetUpdates(offset int, limit int, timeout int, allowedUpdates []string) ([]*Update, error)
}

type telegramBot struct {
    IBot

    httpClient *http.Client
    token string    
}

func (b *telegramBot) queryApi(apiMethod string, params url.Values) ([]byte, error) {
    request, err := http.NewRequest("POST", 
                                    fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, apiMethod),
                                    strings.NewReader(params.Encode()))
    if err != nil {
        return nil, err
    }
    resp, err := b.httpClient.Do(request)
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
        httpClient: &http.Client{},
    }
}

func (b *telegramBot) GetMe() (*User, error) {
    response, err := b.queryApi("getMe", url.Values{})
    if err != nil {
        return nil, err
    }

    responseProto := &GetMeResponse{}
    err = json.Unmarshal(response, responseProto)
    if err != nil {
        return nil, err
    }
    if !responseProto.Ok {
        return nil, fmt.Errorf("Cannot read the result (%d): %s", responseProto.ErrorCode, responseProto.Description)
    }
    return &responseProto.Result, nil
}

func (b *telegramBot) GetUpdates(offset int, limit int, timeout int, allowedUpdates []string) ([]*Update, error) {
    params := url.Values{}
    params.Set("offset", strconv.Itoa(offset))
    params.Set("limit", strconv.Itoa(limit))
    params.Set("timeout", strconv.Itoa(timeout))
    for _, upd := range allowedUpdates {
        params.Add("allowed_updates", upd)
    }
    
    response, err := b.queryApi("getUpdates", params)
    if err != nil {
        return nil, err
    }
    responseProto := &GetUpdateResponse{}
    if err = json.Unmarshal(response, responseProto); err != nil {
        return nil, err
    }
    if !responseProto.Ok {
        return nil, fmt.Errorf("Cannot read the result (%d): %s", responseProto.ErrorCode, responseProto.Description)
    }
    return  responseProto.Result, nil
}
