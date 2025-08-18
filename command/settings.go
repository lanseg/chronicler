package command

import (
	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	"chronicler/adapter/web"
)

type HttpSettings struct {
	CachePath      string `json:"cache_path" env:"CACHE_PATH" arg:"cache_path"`
	CookieJar      string `json:"cookie_jar" env:"COOKIE_JAR" arg:"cookie_jar"`
	RequestDelayMs int    `json:"request_delay_ms" env:"REQUEST_DELAY_MS" arg:"request_delay_ms"`
}

type StorageSettings struct {
	Root string `json:"root" env:"ROOT" arg:"root"`
}

type Settings struct {
	// Adapter settings
	Reddit      *reddit.RedditSettings   `json:"reddit" env:"REDDIT" arg:"reddit"`
	Twitter     *twitter.TwitterSettings `json:"twitter" env:"TWITTER" arg:"twitter"`
	WebSettings *web.WebSettings         `json:"web" env:"WEB" arg:"web"`

	// General settings
	Storage      *StorageSettings `json:"storage" env:"STORAGE" arg:"storage"`
	HttpSettings *HttpSettings    `json:"http" env:"HTTP" arg:"http"`
}
