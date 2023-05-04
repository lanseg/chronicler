package main

import (
	"io/ioutil"
	"web/tokenizer"
)

func main() {
	content, _ := ioutil.ReadFile("sample.html")
	t := tokenizer.CreateTokenizer()
	t.Parse(string(content))
}
