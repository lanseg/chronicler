package command

type HttpSettings struct {
	CachePath      string `json:"cache_path" env:"CACHE_PATH" arg:"cache_path"`
	CookieJar      string `json:"cookie_jar" env:"COOKIE_JAR" arg:"cookie_jar"`
	RequestDelayMs int    `json:"request_delay_ms" env:"REQUEST_DELAY_MS" arg:"request_delay_ms"`
}

type RedditSettings struct {
	Token string `json:"token" env:"TOKEN" arg:"token"`
}

type TwitterSettings struct {
	Token string `json:"token" env:"TOKEN" arg:"token"`
}

type StorageSettings struct {
	Root string `json:"root" env:"ROOT" arg:"root"`
}

type WebSettings struct {
	Recursive bool `json:"recursive" env:"RECURSIVE" arg:"recursive"`
}

type Settings struct {
	Twitter *TwitterSettings `json:"twitter" env:"TWITTER" arg:"twitter"`
	Reddit  *RedditSettings  `json:"reddit" env:"REDDIT" arg:"reddit"`
	Storage *StorageSettings `json:"storage" env:"STORAGE" arg:"storage"`

	WebSettings  *WebSettings  `json:"web" env:"WEB" arg:"web"`
	HttpSettings *HttpSettings `json:"http" env:"HTTP" arg:"http"`
}
