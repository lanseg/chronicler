package reddit

type RedditSettings struct {
	Token string `json:"token" env:"TOKEN" arg:"token"`
}
