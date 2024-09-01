package webdriver

import (
	"net/http"
	"sync"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"
)

type Browser interface {
	RunSession(func(driver WebDriver)) error
}

type fakeBrowser struct {
	Browser

	driver WebDriver
}

func (fwd *fakeBrowser) RunSession(do func(driver WebDriver)) error {
	do(fwd.driver)
	return nil
}

type browserImpl struct {
	Browser

	server string

	logger    *cm.Logger
	scenarios ScenarioLibrary
	driver    WebDriver
	lock      sync.Mutex
}

func (wd *browserImpl) initIfNeeded() error {
	if wd.driver != nil {
		return nil
	}
	driver := NewGeckoDriver(wd.server, &http.Client{})
	sessionId, err := conc.WaitForSomething(driver.NewSession).Get()
	wd.logger.Infof("Started new webdriver session: %s, err: %s", sessionId, err)
	if err != nil {
		return err
	}
	wd.driver = NewScenarioWebdriver(driver, wd.scenarios)
	return nil
}

func (wd *browserImpl) RunSession(session func(driver WebDriver)) error {
	wd.lock.Lock()
	defer wd.lock.Unlock()

	// If not started yet, starting web browser and connecting to it
	if err := wd.initIfNeeded(); err != nil {
		return err
	}
	session(wd.driver)
	return nil
}

func NewFakeBrowser(driver WebDriver) Browser {
	return &fakeBrowser{
		driver: driver,
	}
}

func NewBrowser(webdriverServer string, scenarios string) Browser {
	logger := cm.NewLogger("Browser")
	sc, err := LoadScenarios(scenarios)
	if err != nil {
		logger.Warningf("Unable to load scenarios defined in %q: ", scenarios, err)
	}

	return &browserImpl{
		server:    webdriverServer,
		scenarios: sc,
		logger:    cm.NewLogger("Browser"),
	}
}
