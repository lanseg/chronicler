package telegram;

import (
    "fmt"    
    "io/ioutil"
    "net/http"
    "net/url"
    "encoding/json"
)

type IResponseMetadata interface {
    IsOk() bool
    GetErrorCode() int64
    GetDescription() string
}

type ResponseMetadata struct {
  IResponseMetadata
    
  Ok bool             `json:"ok"`
  ErrorCode int64     `json:"error_code"`
  Description string  `json:"description"`
}

func (dt *ResponseMetadata) IsOk() bool {
    return dt.Ok
}

func (dt *ResponseMetadata) GetErrorCode() int64 {
    return dt.ErrorCode
}

func (dt *ResponseMetadata) GetDescription() string {
    return dt.Description
}

type telegramBot struct {
    httpClient *http.Client
    token string    
}

func NewBot(token string) *telegramBot {
    return &telegramBot {
        token: token,
        httpClient: &http.Client{},
    }
}

func (b *telegramBot) queryApi(apiMethod string, params url.Values) ([]byte, error) {
    resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, apiMethod), params)
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    return body, nil
}

func (b *telegramBot) queryAndUnmarshal(apiMethod string, params url.Values, result interface{}) (interface{}, error) {
    resultBytes, err :=  b.queryApi(apiMethod, params)
    if err != nil {
        return nil, err
    }
    if err = json.Unmarshal(resultBytes, result); err != nil {
        return nil, fmt.Errorf("Cannot unmarshal the response: %s", err)
    }
    resultMeta := result.(IResponseMetadata)
    if !resultMeta.IsOk() {
        return nil, fmt.Errorf("Request \"%s\" completed with error %d: %s", apiMethod, resultMeta.GetErrorCode(), resultMeta.GetDescription())
    }
    
    return result, nil
}

type GetMeResponse struct {
  ResponseMetadata

  Result *User         `json:"result"` 
}

func (b *telegramBot) GetMe() (*User, error) {
    response, err := b.queryAndUnmarshal("getMe", url.Values{}, &GetMeResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetMeResponse).Result, nil
}

type GetUpdatesResponse struct {
  ResponseMetadata

  Result []*Update         `json:"result"` 
}

func (b *telegramBot) GetUpdates(offset int64, limit int64, timeout int64, allowedUpdates []string) ([]*Update, error) {
    params := url.Values{}
    params.Set("chat_id", "-1480532340")
    params.Set("offset", fmt.Sprintf("%d", offset))
    params.Set("limit", fmt.Sprintf("%d", limit))
    params.Set("timeout", fmt.Sprintf("%d", timeout))
    for _, upd := range allowedUpdates {
        params.Add("allowed_updates", upd)
    }
    
    response, err := b.queryAndUnmarshal("getUpdates", params, &GetUpdatesResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetUpdatesResponse).Result, nil
}

type GetFileResponse struct {
  ResponseMetadata
  
  Result *File         `json:"result"`
}

func (b *telegramBot) GetFile(fileId string) (*File, error) {
    params := url.Values{}
    params.Set("file_id", fileId)
    
    response, err := b.queryAndUnmarshal("getFile", params, &GetFileResponse{})
    if err != nil {
        return nil, err
    }
    return response.(*GetFileResponse).Result, nil
}

// Utilities

func GetLargestImage(sizes []*PhotoSize) *PhotoSize {
    if len(sizes) == 0 {
        return nil
    }
    var result *PhotoSize = sizes[0]
    resultSize := int64(0)
    for _, photo := range sizes {
        size := photo.Width * photo.Height
        if size > resultSize {
            result = photo
            resultSize = size
        }
    }
    return result
}
