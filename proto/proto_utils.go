package proto

import (
	"chronicler/common"
	"net/url"
)

func ParsePageLink(page string, link string) (*Link, error) {
	parent, err := url.Parse(page)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if parent.Scheme != "" && url.Scheme == "" {
		url.Scheme = parent.Scheme
	}
	if parent.Host != "" && url.Host == "" {
		url.Host = parent.Host
	}
	if parent.User != nil && url.User == nil {
		url.User = parent.User
	}
	normal := common.NormalizeURL(url)
	if normal == nil || normal.String() == link {
		return &Link{Href: link}, nil
	}
	return &Link{
		Href:     normal.String(),
		Variants: []string{link},
	}, nil
}

func ParseLink(link string) (*Link, error) {
	return ParsePageLink("", link)
}
