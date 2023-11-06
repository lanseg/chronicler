package webdriver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	Match        string `json: match`
	Before       string `json: before`
	beforeScript string
	root         string

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
	if s.re == nil && s.init() != nil {

	}
	return s.re.FindString(url) != ""
}

func (s *ScenarioImpl) BeforeScript() string {
	if s.beforeScript != "" {
		return s.beforeScript
	}
	script, err := os.ReadFile(filepath.Join(s.root, s.Before))
	if err != nil {
		script = []byte(fmt.Sprintf("return false; //%s", err.Error()))
	}
	s.beforeScript = strings.TrimSpace(string(script))
	return s.beforeScript
}

// ScenarioLibary
type ScenarioLibrary interface {
	Matches(url string) Scenario
}

type NoopScenarioLibrary struct {
	ScenarioLibrary
}

func (sl *NoopScenarioLibrary) Matches(url string) Scenario {
	return nil
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
	root := filepath.Dir(path)
	result := []Scenario{}
	for _, s := range scenarios {
		s.root = root
		if err = s.init(); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return &ScenarioLibraryImpl{
		scenarios: result,
	}, nil
}
