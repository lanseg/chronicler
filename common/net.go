package common

import (
	"net/url"
	"path/filepath"
	"strings"
)

func IsSameHost(parent *url.URL, link *url.URL) bool {
	return parent.Hostname() == link.Hostname()
}

func ParseUrlDefaults(link string, defaults *url.URL) (*url.URL, error) {
	result, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if result.Scheme == "" {
		result.Scheme = defaults.Scheme
	}
	if result.Host == "" {
		result.Host = defaults.Host
	}
	if result.Path == "" {
		result.Path = defaults.Path
	}
	return result, nil
}

func NormalizeURL(link *url.URL) *url.URL {
	if link == nil {
		return nil
	}
	port := link.Port()
	if link.Scheme == "" && link.Port() != "443" {
		link.Scheme = "http"
	}
	if (link.Scheme == "http" && port == "80") || (link.Scheme == "https" && port == "443") {
		link.Host = strings.TrimSuffix(link.Host, ":"+port)
	}
	link.Fragment = ""
	link.Host = strings.ToLower(link.Host)
	if link.Path == "" {
		link.Path = "/"
	} else {
		link.Path = filepath.Clean(link.Path)
	}
	link.ForceQuery = false
	link.RawPath = link.Path
	return link
}
