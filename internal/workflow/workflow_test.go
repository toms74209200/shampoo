//go:build small

package workflow

import (
	"crypto/rand"
	"reflect"
	"regexp"
	"testing"
)

func TestParse(t *testing.T) {
	path := rand.Text()
	tests := []struct {
		name    string
		data    string
		want    Workflow
		wantErr bool
	}{
		{name: "when Parse with a pull_request event and paths then returns the workflow expecting those paths", data: "name: go\non:\n  pull_request:\n    paths:\n      - '**.go'\n      - '!vendor/**'\n", want: Workflow{Name: "go", PullRequest: true, paths: []filter{{expr: regexp.MustCompile(`^.*\.go$`)}, {expr: regexp.MustCompile(`^vendor/.*$`), negated: true}}}},
		{name: "when Parse with branch filters then returns the workflow filtering its base", data: "name: go\non:\n  pull_request:\n    branches: [main, 'release/*']\n    branches-ignore: ['wip?']\n    paths-ignore: ['docs/**']\n", want: Workflow{Name: "go", PullRequest: true, branches: []filter{{expr: regexp.MustCompile(`^main$`)}, {expr: regexp.MustCompile(`^release/[^/]*$`)}}, branchesIgnore: []filter{{expr: regexp.MustCompile(`^wip?$`)}}, pathsIgnore: []filter{{expr: regexp.MustCompile(`^docs/.*$`)}}}},
		{name: "when Parse with a pull_request event without filters then returns the workflow expected for every change", data: "name: commit\non:\n  pull_request:\n", want: Workflow{Name: "commit", PullRequest: true}},
		{name: "when Parse with events as a list then returns the workflow on pull requests", data: "name: ci\non: [push, pull_request]\n", want: Workflow{Name: "ci", PullRequest: true}},
		{name: "when Parse with one event then returns the workflow on that event only", data: "name: ci\non: push\n", want: Workflow{Name: "ci"}},
		{name: "when Parse with other events in a mapping then returns the workflow not on pull requests", data: "name: release\non:\n  workflow_dispatch:\n", want: Workflow{Name: "release"}},
		{name: "when Parse with no name then returns the workflow named by its path", data: "on: pull_request\n", want: Workflow{Name: path, PullRequest: true}},
		{name: "when Parse with a broken document then returns an invalid workflow", data: "on: [", wantErr: true},
		{name: "when Parse with filters that are not lists then returns an invalid workflow", data: "on:\n  pull_request:\n    paths: {a: b}\n", wantErr: true},
		{name: "when Parse with a pattern that does not compile then returns an invalid workflow", data: "on:\n  pull_request:\n    paths: ['[a']\n", wantErr: true},
		{name: "when Parse with an empty pattern then returns an invalid workflow", data: "on:\n  pull_request:\n    paths: ['!']\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(path, []byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}

func TestExpects(t *testing.T) {
	goFiles := []filter{{expr: regexp.MustCompile(`^.*\.go$`)}, {expr: regexp.MustCompile(`^vendor/.*$`), negated: true}}
	docs := []filter{{expr: regexp.MustCompile(`^docs/.*$`)}}
	main := []filter{{expr: regexp.MustCompile(`^main$`)}}
	tests := []struct {
		name     string
		workflow Workflow
		base     string
		changed  []string
		want     bool
	}{
		{name: "when Expects with a changed path its paths match then returns true", workflow: Workflow{PullRequest: true, paths: goFiles}, base: "main", changed: []string{"README.md", "cmd/main.go"}, want: true},
		{name: "when Expects with changed paths its paths do not match then returns false", workflow: Workflow{PullRequest: true, paths: goFiles}, base: "main", changed: []string{"README.md"}, want: false},
		{name: "when Expects with a changed path a later negated pattern matches then returns false", workflow: Workflow{PullRequest: true, paths: goFiles}, base: "main", changed: []string{"vendor/lib.go"}, want: false},
		{name: "when Expects with every changed path ignored then returns false", workflow: Workflow{PullRequest: true, pathsIgnore: docs}, base: "main", changed: []string{"docs/a.md"}, want: false},
		{name: "when Expects with a changed path not ignored then returns true", workflow: Workflow{PullRequest: true, pathsIgnore: docs}, base: "main", changed: []string{"docs/a.md", "main.go"}, want: true},
		{name: "when Expects with a base its branches do not match then returns false", workflow: Workflow{PullRequest: true, branches: main}, base: "develop", changed: []string{"main.go"}, want: false},
		{name: "when Expects with a base its branches match then returns true", workflow: Workflow{PullRequest: true, branches: main}, base: "main", changed: []string{"main.go"}, want: true},
		{name: "when Expects with an ignored base then returns false", workflow: Workflow{PullRequest: true, branchesIgnore: main}, base: "main", changed: []string{"main.go"}, want: false},
		{name: "when Expects with no filters then returns true", workflow: Workflow{PullRequest: true}, base: "main", changed: []string{"main.go"}, want: true},
		{name: "when Expects with a workflow not on pull requests then returns false", workflow: Workflow{}, base: "main", changed: []string{"main.go"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.workflow.Expects(tt.base, tt.changed)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpected(t *testing.T) {
	tests := []struct {
		name    string
		sources []Source
		base    string
		changed []string
		want    []string
		wantErr bool
	}{
		{name: "when Expected with workflows of matching and missing paths then returns the matching names", sources: []Source{{Path: "a.yml", Data: []byte("name: go\non:\n  pull_request:\n    paths: ['**.go']\n")}, {Path: "b.yml", Data: []byte("name: docs\non:\n  pull_request:\n    paths: ['docs/**']\n")}}, base: "master", changed: []string{"main.go"}, want: []string{"go"}},
		{name: "when Expected with no workflows then returns none", sources: nil, base: "master", changed: []string{"main.go"}, want: nil},
		{name: "when Expected with an invalid workflow then returns an error", sources: []Source{{Path: "a.yml", Data: []byte("on: [")}}, base: "master", changed: []string{"main.go"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expected(tt.sources, tt.base, tt.changed)

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %v, %v, want %v", got, err, tt.want)
			}
		})
	}
}
