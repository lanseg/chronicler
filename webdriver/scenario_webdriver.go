package webdriver

import (
	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"
)

type scenarioWebdriver struct {
	WebDriver

	url        string
	logger     *cm.Logger
	baseDriver WebDriver
	scenarios  ScenarioLibrary
}

func isNotEmpty(s string) bool {
	return s != ""
}

func (wd *scenarioWebdriver) getScenario() optional.Optional[Scenario] {
	scenario := wd.scenarios.Matches(wd.url)
	if scenario == nil {
		return optional.Nothing[Scenario]{}
	}
	wd.logger.Debugf("Found matching scenario for %s", wd.url)
	return optional.Of(scenario)
}

func (wd *scenarioWebdriver) runScenario(script string) {
	concurrent.WaitForValueRetries("true", func() optional.Optional[string] {
		return wd.baseDriver.ExecuteScript(script)
	}, 120)
}

func (wd *scenarioWebdriver) NewSession() {
}

func (wd *scenarioWebdriver) Navigate(url string) {
	wd.baseDriver.Navigate(url)
	wd.url = url

	optional.MapErr(wd.getScenario(), func(s Scenario) (string, error) {
		return s.BeforeAll()
	}).
		Filter(isNotEmpty).
		IfPresent(wd.runScenario)
}

func (wd *scenarioWebdriver) GetPageSource() optional.Optional[string] {
	return wd.baseDriver.GetPageSource()
}

func (wd *scenarioWebdriver) GetCurrentURL() optional.Optional[string] {
	currentUrl := wd.baseDriver.GetCurrentURL()
	currentUrl.IfPresent(func(url string) {
		wd.url = url
	})
	return currentUrl
}

func (wd *scenarioWebdriver) Print() optional.Optional[string] {
	optional.MapErr(wd.getScenario(), func(s Scenario) (string, error) {
		return s.BeforePrint()
	}).
		Filter(isNotEmpty).
		IfPresent(wd.runScenario)
	return wd.baseDriver.Print()
}

func (wd *scenarioWebdriver) TakeScreenshot() optional.Optional[string] {
	return wd.baseDriver.TakeScreenshot()
}

func (wd *scenarioWebdriver) ExecuteScript(script string) optional.Optional[string] {
	return wd.baseDriver.ExecuteScript(script)
}

func NewScenarioWebdriver(wd WebDriver, scenarios ScenarioLibrary) WebDriver {
	return &scenarioWebdriver{
		logger:     cm.NewLogger("ScenarioWebdriver"),
		baseDriver: wd,
		scenarios:  scenarios,
	}
}
