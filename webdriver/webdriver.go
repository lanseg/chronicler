package webdriver

import (
	"github.com/lanseg/golang-commons/optional"
)

type WebDriver interface {
	NewSession()
	Navigate(string)
	GetPageSource() optional.Optional[string]
	TakeScreenshot() optional.Optional[string]
	Print() optional.Optional[string]
	ExecuteScript(string) optional.Optional[string]
	SetScenarios(ScenarioLibrary)
}

type NoopWebdriver struct {
	WebDriver
}

func (*NoopWebdriver) NewSession()     {}
func (*NoopWebdriver) Navigate(string) {}
func (*NoopWebdriver) GetPageSource() optional.Optional[string] {
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
func (*NoopWebdriver) SetScenarios(ScenarioLibrary) {
}
