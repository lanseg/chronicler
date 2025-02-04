package fourchan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type FourChanResponse struct {
	Posts []*FourChanPost `json:"posts"`
}

type FourChanPost struct {
	No            int64  `json:"no"`
	Resto         int64  `json:"resto"`
	Sticky        int64  `json:"sticky"`
	Closed        int64  `json:"closed"`
	Now           string `json:"now"`
	Time          int64  `json:"time"`
	Name          string `json:"name"`
	Trip          string `json:"trip"`
	Id            string `json:"id"`
	CapCode       string `json:"capcode"`
	Country       string `json:"country"`
	CountryName   string `json:"country_name"`
	BoardFlag     string `json:"board_flag"`
	FlagName      string `json:"flag_name"`
	Sub           string `json:"sub"`
	Com           string `json:"com"`
	Tim           int64  `json:"tim"`
	Filename      string `json:"filename"`
	Ext           string `json:"ext"`
	FSize         int64  `json:"fsize"`
	MD5           string `json:"md5"`
	W             int64  `json:"w"`
	H             int64  `json:"h"`
	Tnw           int64  `json:"tn_w"`
	Tnh           int64  `json:"tn_h"`
	FileDeleted   int64  `json:"filedeleted"`
	Spoiler       int64  `json:"spoiler"`
	CustomSpoiler int64  `json:"custom_spoiler"`
	Replies       int64  `json:"replies"`
	Images        int64  `json:"images"`
	BumpLimit     int64  `json:"bumplimit"`
	ImageLimit    int64  `json:"imagelimit"`
	Tag           string `json:"tag"`
	SemanticUrl   string `json:"semantic_url"`
	Since4Pass    int64  `json:"since4pass"`
	UniqueIps     int64  `json:"unique_ips"`
	MImg          int64  `json:"m_img"`
	Archived      int64  `json:"archived"`
	ArchivedOn    int64  `json:"archived_on"`
}

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func GetThread(httpClient HttpClient, board string, thread string) ([]*FourChanPost, error) {
	requestUrl, err := url.Parse(fmt.Sprintf("https://a.4cdn.org/%s/thread/%s.json", board, thread))
	if err != nil {
		return nil, err
	}

	request := &http.Request{Method: "GET", URL: requestUrl}
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rp := &FourChanResponse{}
	if err = json.Unmarshal(data, rp); err != nil {
		return nil, err
	}

	return rp.Posts, nil
}
