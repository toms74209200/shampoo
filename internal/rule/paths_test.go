//go:build small

package rule

import (
	"regexp"
	"testing"
)

func TestCovers(t *testing.T) {
	tests := []struct {
		name    string
		paths   Paths
		changed []string
		want    bool
	}{
		{name: "when Covers with no patterns and a changed path then returns true", paths: Paths{}, changed: []string{"main.go"}, want: true},
		{name: "when Covers with no patterns and no changed path then returns false", paths: Paths{}, changed: nil, want: false},
		{name: "when Covers with a pattern matching a changed path then returns true", paths: Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`)}}, changed: []string{"README.md", "internal/rule/rule.go"}, want: true},
		{name: "when Covers with no pattern matching a changed path then returns false", paths: Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`)}}, changed: []string{"README.md"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.paths.Covers(tt.changed)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
