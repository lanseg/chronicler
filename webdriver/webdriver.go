package webdriver

import (
	"chronicler/util"
	"sync"
	"time"

	"github.com/lanseg/golang-commons/optional"
)

const (
	browserProfileFolder = "/tmp/tmp.QTFqrzeJX4/"
	webdriverPort        = 2828
	webdriverAddress     = "127.0.0.1"
	connectRetries       = 10
	connectRetryInterval = 3 * time.Second
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

type ExclusiveWebDriver struct {
	driver WebDriver
	mu     sync.Mutex
}

func (e *ExclusiveWebDriver) Batch(do func(driver WebDriver)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	do(e.driver)
}

func Connect() optional.Optional[*ExclusiveWebDriver] {
	driver, err := connectMarionette(webdriverAddress, webdriverPort).Get()
	if err == nil {
		return optional.OfNullable(&ExclusiveWebDriver{
			driver: driver,
		})
	}

	startFirefox(webdriverPort, browserProfileFolder)

	return optional.Map(
		util.WaitForPresent(func() optional.Optional[WebDriver] {
			return connectMarionette(webdriverAddress, webdriverPort)
		}, connectRetries, connectRetryInterval), func(driver WebDriver) *ExclusiveWebDriver {
			return &ExclusiveWebDriver{
				driver: driver,
			}
		})
}
