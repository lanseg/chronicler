package command

type HttpSettings struct {
	CachePath string `json:"cache_path" env:"CACHE_PATH" arg:"cache_path"`
	CookieJar string `json:"cookie_jar" env:"COOKIE_JAR" arg:"cookie_jar"`
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

type Settings struct {
	Twitter *TwitterSettings `json:"twitter" env:"TWITTER" arg:"twitter"`
	Reddit  *RedditSettings  `json:"reddit" env:"REDDIT" arg:"reddit"`
	Storage *StorageSettings `json:"storage" env:"STORAGE" arg:"storage"`

	HttpSettings *HttpSettings `json:"http" env:"HTTP" arg:"http"`
}
