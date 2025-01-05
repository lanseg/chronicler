package webdriver

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestScenarios(t *testing.T) {

	scenarios := &ScenarioLibraryImpl{
		scenarios: []Scenario{},
	}
	for _, sci := range []*ScenarioImpl{
		{
			config: ScenarioConfig{
				Match:     "some.*match",
				BeforeAll: "beforescript.js",
			},
		},
		{
			config: ScenarioConfig{
				Match:     "^http://exact.site/match$",
				BeforeAll: "beforescript.js",
			},
		},
		{
			config: ScenarioConfig{
				Match:     "noRegexpMatch",
				BeforeAll: "beforescript.js",
			},
		},
	} {
		sci.root = "testdata"
		scenarios.scenarios = append(scenarios.scenarios, sci)
	}

	for _, tc := range []struct {
		name string
		url  string
		want string
	}{
		{
			name: "no url matches",
			url:  "",
			want: "",
		},
		{
			name: "single scenario matches",
			url:  "http://exact.site/match",
			want: "return true; //before",
		},
		{
			name: "multiple scenarios match",
			url:  "match",
			want: "",
		},
		{
			name: "A longer no regexp match",
			url:  "a longer noRegexpMatch some tail",
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mch := scenarios.Matches(tc.url)
			if mch == nil && tc.want == "" {
				return
			}

			ba, err := mch.BeforeAll()
			if err != nil {
				t.Errorf("Unexpected error %s", err)
				return
			}

			if ba != tc.want {
				t.Errorf("Expected to match %v, but got %v", tc.want, mch)
			}
		})
	}
}

func TestScenarioLoad(t *testing.T) {
	for _, tc := range []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name: "normal scenario",
			file: "normal_scenarios.json",
		},
		{
			name:    "malformed json",
			file:    "malformed.json",
			wantErr: true,
		},
		{
			name:    "malformed regexp",
			file:    "malformed_regexp.json",
			wantErr: true,
		},
		{
			name: "multiline scenario",
			file: "multiline_scenarios.json",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc, err := LoadScenarios(filepath.Join("testdata", tc.file))
			if err == nil && tc.wantErr {
				t.Errorf("No expected error")
			}
			if err != nil && !tc.wantErr {
				t.Errorf("Unexpected error %s", err)
			}
			fmt.Println(sc)
		})
	}
}
