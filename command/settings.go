package command

import "os"

type HttpSettings struct {
	CachePath string `json:"cache_path"`
	CookieJar string `json:"cookie_jar"`
}
type RedditSettings struct {
	Token string `json:"token"`
}
type StorageSettings struct {
	Root string `json:"root"`
}
type TwitterSettings struct {
	Token string `json:"token"`
}

type Settings struct {
	Twitter *TwitterSettings `json:"twitter"`
	Reddit  *RedditSettings  `json:"reddit"`
	Storage *StorageSettings `json:"storage"`

	HttpSettings *HttpSettings `json:"http"`
}

func GetSettings() *Settings {
	return &Settings{
		Twitter: &TwitterSettings{
			Token: os.Getenv("TWITTER_TOKEN"),
		},
		Reddit: &RedditSettings{
			Token: os.Getenv("REDDIT_TOKEN"),
		},
		Storage: &StorageSettings{
			Root: os.Getenv("STORAGE_ROOT"),
		},
		HttpSettings: &HttpSettings{
			CachePath: os.Getenv("CACHE_PATH"),
			CookieJar: os.Getenv("COOKIE_JAR"),
		},
	}
}
