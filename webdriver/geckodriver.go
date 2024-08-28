package webdriver

import (
	"fmt"
	"os"

	cm "github.com/lanseg/golang-commons/common"
	opt "github.com/lanseg/golang-commons/optional"
)

type Response[T any] struct {
	Value T `json:"value"`
}

func getValue[T any](r *Response[T]) T {
	return r.Value
}

type Capabilities struct{}

type CreateSession struct {
	DesiredCapabilities Capabilities `json:"desiredCapabilities"`
}

type MaybeError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Stacktrace string `json:"stacktrace"`
}

type Session struct {
	MaybeError

	SessionId string `json:"sessionId"`
}

type GeckoDriver struct {
	WebDriver

	session *Session
	url     string
	client  HttpClient
	logger  *cm.Logger
}

func (gd *GeckoDriver) command(command string) string {
	return fmt.Sprintf("%s/session/%s/%s", gd.url, gd.session.SessionId, command)
}

func (gd *GeckoDriver) NewSession() opt.Optional[string] {
	return opt.MapErr(
		opt.OfError(NewTypedRequestBuilder[Response[*Session]](gd.url+"/session").
			WithMethod("POST").
			WithJsonBody(&CreateSession{}).
			DoAndUnmarshal(gd.client)),
		func(r *Response[*Session]) (string, error) {
			if r.Value.Error != "" {
				return "", fmt.Errorf("Error %q: %s", r.Value.Error, r.Value.Message)
			}
			gd.session = r.Value
			return r.Value.SessionId, nil
		})
}

func (gd *GeckoDriver) Navigate(url string) {
	NewRequestBuilder(gd.command("url")).
		WithMethod("POST").
		WithJsonBody(struct {
			Url string `json:"url"`
		}{Url: url}).
		Do(gd.client)
}

func (gd *GeckoDriver) GetPageSource() opt.Optional[string] {
	return opt.Map(
		opt.OfError(
			NewTypedRequestBuilder[Response[string]](gd.command("source")).
				DoAndUnmarshal(gd.client)), getValue)
}

func (gd *GeckoDriver) GetCurrentURL() opt.Optional[string] {
	return opt.Map(
		opt.OfError(
			NewTypedRequestBuilder[Response[string]](gd.command("url")).DoAndUnmarshal(gd.client)),
		getValue)
}

func (gd *GeckoDriver) TakeScreenshot() opt.Optional[string] {
	return opt.Map(
		opt.OfError(
			NewTypedRequestBuilder[Response[string]](gd.command("moz/screenshot/full")).
				DoAndUnmarshal(gd.client)),
		getValue)
}

func (gd *GeckoDriver) Print() opt.Optional[string] {
	return opt.Map(opt.OfError(
		NewTypedRequestBuilder[Response[string]](
			gd.command("print")).
			WithMethod("POST").
			WithJsonBody(struct{}{}).
			DoAndUnmarshal(gd.client)), getValue)
}

func (gd *GeckoDriver) ExecuteScript(script string) opt.Optional[string] {
	return opt.Map(opt.OfError(
		NewTypedRequestBuilder[Response[string]](
			gd.command("execute/sync")).
			WithMethod("POST").
			WithJsonBody(struct {
				Script string   `json:"script"`
				Args   []string `json:"args"`
			}{
				Script: script,
				Args:   []string{},
			}).
			DoAndUnmarshal(gd.client),
	), getValue)
}

func (gd *GeckoDriver) Doit() {
	gd.NewSession().If(func(s string) {
		fmt.Printf("SESSION: %s\n", s)
	}, func(e error) {
		fmt.Printf("ERR: %s\n", e)
		os.Exit(-1)
	}, nil)
	gd.Navigate("https://google.com")
	gd.GetCurrentURL().IfPresent(func(s string) {
		fmt.Printf("CURRENT URL: %s\n", s)
	})
	gd.GetPageSource().IfPresent(func(s string) {
		fmt.Printf("PAGE CONTENTS: %s...\n", s[:100])
	})
	gd.Print().IfPresent(func(s string) {
		fmt.Printf("PRINT %s...\n", s[:100])
	})
	gd.TakeScreenshot().IfPresent(func(s string) {
		fmt.Printf("SCREENSHOT %s...\n", s[:100])
	})
	gd.ExecuteScript("return 'hello world';").If(
		func(s string) {
			fmt.Printf("SCRIPT: %s\n", s)
		}, func(e error) {
			fmt.Printf("ERROR: %s\n", e)
		}, nil)
}

func NewGeckoDriver(url string, client HttpClient) WebDriver {
	return &GeckoDriver{
		url:    url,
		logger: cm.NewLogger("GeckoDriver"),
		client: client,
	}
}
