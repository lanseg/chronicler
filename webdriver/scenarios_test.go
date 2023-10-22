package webdriver

import (
	"testing"
)

func TestScenarios(t *testing.T) {

	scenarios := &ScenarioLibraryImpl{
		scenarios: []Scenario{},
	}
	for _, sci := range []*ScenarioImpl{
		{
			Match:  "some.*match",
			Before: "Before path",
		},
		{
			Match:  "^http://exact.site/match$",
			Before: "Exact match",
		},
		{
			Match:  "noRegexpMatch",
			Before: "No regexp match",
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
