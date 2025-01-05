package webdriver

import (
	"time"

	"github.com/lanseg/golang-commons/optional"
)

const (
	connectRetries       = 10
	connectRetryInterval = 3 * time.Second
)

type WebDriver interface {
	NewSession() optional.Optional[string]

	Navigate(string)
	GetPageSource() optional.Optional[string]
	GetCurrentURL() optional.Optional[string]
	Print() optional.Optional[string]
	TakeScreenshot() optional.Optional[string]

	ExecuteScript(string) optional.Optional[string]
}

type NoopWebdriver struct {
	WebDriver
}

func (*NoopWebdriver) NewSession() optional.Optional[string] {
	return optional.Of("Session id")
}

func (*NoopWebdriver) Navigate(string) {}
func (*NoopWebdriver) GetPageSource() optional.Optional[string] {
	return optional.Of("")
}
func (*NoopWebdriver) GetCurrentURL() optional.Optional[string] {
	return optional.Of("")
}
func (*NoopWebdriver) TakeScreenshot() optional.Optional[string] {
	return optional.Of("")
}
func (*NoopWebdriver) Print() optional.Optional[string] {
	return optional.Of("")
}
func (*NoopWebdriver) ExecuteScript(string) optional.Optional[string] {
	return optional.Of("")
}
