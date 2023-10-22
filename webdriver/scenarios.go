package webdriver

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Scenario
type Scenario interface {
	Matches(url string) bool
	BeforeScript() string
}

type ScenarioImpl struct {
	Scenario

	Match  string   `json: match`
	Before []string `json: before`

	re *regexp.Regexp
}

func (s *ScenarioImpl) init() error {
	re, err := regexp.CompilePOSIX(fmt.Sprintf("^%s$", s.Match))
	if err != nil {
		return err
	}
	s.re = re
	return nil
}

func (s *ScenarioImpl) Matches(url string) bool {
	if s.re == nil {
		s.init()
	}
	return s.re.FindString(url) != ""
}

func (s *ScenarioImpl) BeforeScript() string {
	return strings.Join(s.Before, "\n")
}

// ScenarioLibary
type ScenarioLibrary interface {
	Matches(url string) Scenario
}

type ScenarioLibraryImpl struct {
	ScenarioLibrary

	scenarios []Scenario
}

func (sl *ScenarioLibraryImpl) Matches(url string) Scenario {
	for _, sc := range sl.scenarios {
		if sc.Matches(url) {
			return sc
		}
	}
	return nil
}

// Loading
func LoadScenarios(path string) (ScenarioLibrary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	scenarios := []*ScenarioImpl{}
	err = json.Unmarshal(data, &scenarios)
	if err != nil {
		return nil, err
	}

	result := []Scenario{}
	for _, s := range scenarios {
		result = append(result, s)
	}
	return &ScenarioLibraryImpl{
		scenarios: result,
	}, nil
}
