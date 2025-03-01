package common

import "net/url"

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
