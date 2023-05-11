package main

import (
	"fmt"
	"io/ioutil"
	"web/tokenizer"
)

func main() {
	content, _ := ioutil.ReadFile("sample.html")
	tokens := tokenizer.Tokenize(string(content))
	for i, t := range tokens {
		fmt.Printf("%03d %s\n", i, t)
	}
}
