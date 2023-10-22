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
			Match:  "some.*match",
			Before: []string{"Before path"},
		},
		{
			Match:  "^http://exact.site/match$",
			Before: []string{"Exact match"},
		},
		{
			Match:  "noRegexpMatch",
			Before: []string{"No regexp match"},
		},
	} {
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
			want: "Exact match",
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
			if (mch == nil && tc.want != "") || (mch != nil && mch.BeforeScript() != tc.want) {
				t.Errorf("Expected to match %s, but got %s", tc.want, mch)
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
			if (err != nil && !tc.wantErr) || (err == nil && tc.wantErr) {
				t.Errorf("Unexpected error or no error when expected: %s", err)
			}
			if sc != nil {
				fmt.Println(sc)
			}
		})
	}
}
