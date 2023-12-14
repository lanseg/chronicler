package webdriver

import (
	"sync"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"
)

const (
	browserProfileFolder = "/tmp/tmp.QTFqrzeJX4/"
	webdriverPort        = 2828
	webdriverAddress     = "127.0.0.1"
)

type WebdriverService interface {
	RunSession(func(driver WebDriver)) error
}

type fakeWebdriverService struct {
	WebdriverService

	driver WebDriver
}

func (fwd *fakeWebdriverService) RunSession(do func(driver WebDriver)) error {
	do(fwd.driver)
	return nil
}

type webdriverServiceImpl struct {
	WebdriverService

	logger    *cm.Logger
	scenarios ScenarioLibrary
	driver    WebDriver
	lock      sync.Mutex
}

func (wd *webdriverServiceImpl) initIfNeeded() error {
	if wd.driver != nil {
		return nil
	}

	driver, err := connectMarionette(webdriverAddress, webdriverPort).Get()
	if err == nil {
		driver.NewSession()
		wd.driver = NewScenarioWebdriver(driver, wd.scenarios)
		return nil
	}

	startFirefox(webdriverPort, browserProfileFolder)
	driver, err = concurrent.WaitForSomething(func() optional.Optional[WebDriver] {
		return connectMarionette(webdriverAddress, webdriverPort)
	}).Get()
	if err != nil {
		return err
	}

	driver.NewSession()
	wd.driver = NewScenarioWebdriver(driver, wd.scenarios)
	wd.driver.NewSession()
	return nil
}

func (wd *webdriverServiceImpl) RunSession(session func(driver WebDriver)) error {
	wd.lock.Lock()
	defer wd.lock.Unlock()

	// If not started yet, starting web browser and connecting to it
	if err := wd.initIfNeeded(); err != nil {
		return err
	}
	session(wd.driver)
	return nil
}

func NewFakeWebdriverService(driver WebDriver) WebdriverService {
	return &fakeWebdriverService{
		driver: driver,
	}
}

func NewWebdriverService(scenarios string) WebdriverService {
	logger := cm.NewLogger("WebdriverService")
	sc, err := LoadScenarios(scenarios)
	if err != nil {
		logger.Warningf("Unable to load scenarios defined in %q", scenarios)
	}

	return &webdriverServiceImpl{
		scenarios: sc,
		logger:    cm.NewLogger("WebdriverService"),
	}
}
