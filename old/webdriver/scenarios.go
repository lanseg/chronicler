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
	BeforeAll() (string, error)
	BeforePrint() (string, error)
}

type ScenarioConfig struct {
	Match       string `json: match`
	BeforeAll   string `json: beforeAll`
	BeforePrint string `json: beforePrint`
}

type ScenarioImpl struct {
	Scenario

	root    string
	config  ScenarioConfig
	scripts map[string]string
	match   *regexp.Regexp
}

func (s *ScenarioImpl) init() error {
	re, err := regexp.CompilePOSIX(fmt.Sprintf("^%s$", s.config.Match))
	if err != nil {
		return err
	}
	s.match = re
	return nil
}

func (s *ScenarioImpl) getContent(name string) (string, error) {
	if s.scripts == nil {
		s.scripts = map[string]string{}
	}
	if content, ok := s.scripts[name]; ok {
		return content, nil
	}
	scriptBytes, err := os.ReadFile(filepath.Join(s.root, name))
	if err != nil {
		return "", err
	}
	s.scripts[name] = strings.TrimSpace(string(scriptBytes))
	return s.scripts[name], nil

}

func (s *ScenarioImpl) Matches(url string) bool {
	if s.match == nil && s.init() != nil {
		return false
	}
	return s.match.FindString(url) != ""
}

func (s *ScenarioImpl) BeforeAll() (string, error) {
	if s.config.BeforeAll != "" {
		return s.getContent(s.config.BeforeAll)
	}
	return "", nil
}

func (s *ScenarioImpl) BeforePrint() (string, error) {
	if s.config.BeforePrint != "" {
		return s.getContent(s.config.BeforePrint)
	}
	return "", nil
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
	configs := []*ScenarioConfig{}
	err = json.Unmarshal(data, &configs)
	if err != nil {
		return nil, err
	}
	root := filepath.Dir(path)
	result := []Scenario{}
	for _, cfg := range configs {
		newScenario := &ScenarioImpl{
			config: *cfg,
			root:   root,
		}
		if err = newScenario.init(); err != nil {
			return nil, err
		}
		result = append(result, newScenario)
	}
	return &ScenarioLibraryImpl{
		scenarios: result,
	}, nil
}
