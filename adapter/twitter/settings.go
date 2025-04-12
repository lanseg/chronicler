package twitter

type Settings struct {
	Token string `json:"token" env:"TWITTER_TOKEN"`
}
