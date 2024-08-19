package webdriver

import (
	"sync"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"
)

const (
	browserProfileFolder = "/tmp/tmp.QTFqrzeJX4/"
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
	driver, err := connectMarionette(wd.server).Get()
	if err == nil {
		driver.NewSession()
		wd.driver = NewScenarioWebdriver(driver, wd.scenarios)
		return nil
	}
	driver, err = concurrent.WaitForSomething(func() optional.Optional[WebDriver] {
		return connectMarionette(wd.server)
	}).Get()
	if err != nil {
		return err
	}

	driver.NewSession()
	wd.driver = NewScenarioWebdriver(driver, wd.scenarios)
	wd.driver.NewSession()
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
		logger.Warningf("Unable to load scenarios defined in %q", scenarios)
	}

	return &browserImpl{
		server:    webdriverServer,
		scenarios: sc,
		logger:    cm.NewLogger("Browser"),
	}
}
