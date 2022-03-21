package telegram

import (
    "strconv"
    "net/url"
)


type GetMeResponse struct {
  ResponseMetadata

  Result *User         `json:"result"` 
}

type GetUpdateResponse struct {
  ResponseMetadata

  Result []*Update     `json:"result"` 
}

type GetChatResponse struct {
  ResponseMetadata
  Result *Chat `json:"result"`
}

func (b *telegramBot) GetMe() (*User, error) {
    response, err := b.queryAndUnmarshal("getMe", url.Values{}, &GetMeResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetMeResponse).Result, nil
}

func (b *telegramBot) GetChat(chatId string) (*Chat, error) {
    params := url.Values{}
    params.Set("chat_id", chatId)
    
    response, err := b.queryAndUnmarshal("getChat", params, &GetChatResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetChatResponse).Result, nil
}

func (b *telegramBot) GetUpdates(offset int, limit int, timeout int, allowedUpdates []string) ([]*Update, error) {
    params := url.Values{}
    params.Set("offset", strconv.Itoa(offset))
    params.Set("limit", strconv.Itoa(limit))
    params.Set("timeout", strconv.Itoa(timeout))
    for _, upd := range allowedUpdates {
        params.Add("allowed_updates", upd)
    }
    
    response, err := b.queryAndUnmarshal("getUpdates", params, &GetUpdateResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetUpdateResponse).Result, nil
}

