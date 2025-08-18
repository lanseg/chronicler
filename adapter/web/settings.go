package web

type WebSettings struct {
	Recursive bool `json:"recursive" env:"RECURSIVE" arg:"recursive"`
}
