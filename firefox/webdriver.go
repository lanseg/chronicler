package firefox

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
}

type FakeWebDriver struct {
	WebDriver
}

func (f *FakeWebDriver) NewSession()     {}
func (f *FakeWebDriver) Navigate(string) {}
func (f *FakeWebDriver) GetPageSource() optional.Optional[string] {
	return optional.Of("")
}
func (f *FakeWebDriver) TakeScreenshot() optional.Optional[string] {
	return optional.Of("")
}
func (f *FakeWebDriver) Print() optional.Optional[string] {
	return optional.Of("")
}
func (f *FakeWebDriver) ExecuteScript(string) optional.Optional[string] {
	return optional.Of("")
}
