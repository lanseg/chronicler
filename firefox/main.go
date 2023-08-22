package main

import (
	"encoding/base64"
	"os"
	"sync"

	"chronicler/firefox"
)

func saveBase64(fname string) func(string) {
	return func(content string) {
		sDec, _ := base64.StdEncoding.DecodeString(content)
		os.WriteFile(fname, sDec, 0666)
	}
}

func main() {
	ff := firefox.StartFirefox(2828, "/tmp/tmp.QTFqrzeJX4/")
	defer ff.Driver.Close()

	mn := ff.Driver
	mn.NewSession()
	mn.Navigate("https://meduza.io/")
	mn.TakeScreenshot().IfPresent(saveBase64("/home/lans/devel/chronist/screenshot.png"))
	mn.Print().IfPresent(saveBase64("/home/lans/devel/chronist/pagepdf.pdf"))

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
