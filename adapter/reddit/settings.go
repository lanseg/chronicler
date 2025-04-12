package reddit

type Settings struct {
	Token string `json:"token" env:"REDDIT_TOKEN"`
}
