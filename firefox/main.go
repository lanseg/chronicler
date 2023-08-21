package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"chronicler/firefox"
)

func saveBase64(fname string) func(string) {
	return func(content string) {
		sDec, _ := base64.StdEncoding.DecodeString(content)
		os.WriteFile(fname, sDec, 0666)
	}
}

func main() {
	go firefox.StartFirefox(2828, "/tmp/tmp.QTFqrzeJX4/")
	time.Sleep(10 * time.Second)

	mn, err := firefox.ConnectMarionette("127.0.0.1", 2828).Get()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mn.Close()

	mn.NewSession()
	mn.Navigate("https://meduza.io/")
	mn.TakeScreenshot().IfPresent(saveBase64("/home/lans/devel/chronist/screenshot.png"))
	mn.Print().IfPresent(saveBase64("/home/lans/devel/chronist/pagepdf.pdf"))

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
