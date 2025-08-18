package twitter

type TwitterSettings struct {
	Token string `json:"token" env:"TOKEN" arg:"token"`
}
