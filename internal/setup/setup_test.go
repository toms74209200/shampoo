//go:build small

package setup

import (
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/rule"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Plan
		wantErr bool
	}{
		{name: "when Parse with checks of recommended rules then returns them in order with the default size", data: `{"tasks": "todo.md", "checks": {"small-tests": {"command": "make test"}, "lint": {"command": "make lint", "size": "medium", "paths": ["**.go"], "exceptions": "out/lint.txt"}}}`, want: Plan{Tasks: "todo.md", Checks: []Check{{Rule: "lint", Command: "make lint", Size: rule.Size("medium"), Paths: []string{"**.go"}, Exceptions: "out/lint.txt"}, {Rule: "small-tests", Command: "make test", Size: rule.Size("small")}}}},
		{name: "when Parse with no checks then returns the task source only", data: `{"tasks": "todo.md"}`, want: Plan{Tasks: "todo.md"}},
		{name: "when Parse with a rule that is not recommended then returns an error", data: `{"checks": {"style": {"command": "make style"}}}`, wantErr: true},
		{name: "when Parse with an empty command then returns an error", data: `{"checks": {"lint": {"command": ""}}}`, wantErr: true},
		{name: "when Parse with exceptions of another check than lint then returns an error", data: `{"checks": {"coverage": {"command": "make cover", "exceptions": "out/x.txt"}}}`, wantErr: true},
		{name: "when Parse with an unknown size then returns an error", data: `{"checks": {"lint": {"command": "make lint", "size": "huge"}}}`, wantErr: true},
		{name: "when Parse with an unknown field then returns an error", data: `{"task": "todo.md"}`, wantErr: true},
		{name: "when Parse with invalid JSON then returns an error", data: `{`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data))

			if (err != nil) != tt.wantErr || (!tt.wantErr && !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v, error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestFiles(t *testing.T) {
	tests := []struct {
		name     string
		plan     Plan
		path     string
		want     string
		wantSize int
	}{
		{name: "when Files with a lint check then returns its rule with its paths and exceptions", plan: Plan{Tasks: "todo.md", Checks: []Check{{Rule: "lint", Command: "make lint", Size: rule.Size("small"), Paths: []string{"**.go"}, Exceptions: "out/lint.txt"}}}, path: ".shampoo/rules/lint.toml", want: "rule = \"The code passes the project's linter.\"\ncheck = \"lint\"\ntransitions = [\"edit\", \"task\", \"pr\"]\npaths = [\"**.go\"]\n\n[judge]\n  [judge.exit]\n    1 = \"task-failure\"\n    other = \"workflow-failure\"\n\n[exceptions]\n  file = \"out/lint.txt\"\n", wantSize: 13},
		{name: "when Files with a check then returns the check with its command and size", plan: Plan{Tasks: "todo.md", Checks: []Check{{Rule: "coverage", Command: "make cover", Size: rule.Size("medium")}}}, path: ".shampoo/checks/coverage.toml", want: "command = \"make cover\"\nsize = \"medium\"\n", wantSize: 13},
		{name: "when Files with a commit message check then returns its rule guarding pr only", plan: Plan{Tasks: "todo.md", Checks: []Check{{Rule: "commit-message", Command: "check <commit>", Size: rule.Size("small")}}}, path: ".shampoo/rules/commit-message.toml", want: "rule = \"Every commit message follows the project's format.\"\ncheck = \"commit-message\"\ntransitions = [\"pr\"]\n\n[judge]\n  [judge.exit]\n    1 = \"task-failure\"\n    other = \"workflow-failure\"\n", wantSize: 13},
		{name: "when Files with no checks then returns the settings with the task source", plan: Plan{Tasks: "todo.md"}, path: ".shampoo/settings.toml", want: "[orchestrator]\n  tasks = \"todo.md\"\n", wantSize: 11},
		{name: "when Files with no checks then returns the recommended policies for review", plan: Plan{Tasks: "todo.md"}, path: ".shampoo/policies/dead-code.toml", want: "review = true\ntext = \"No code, argument or setting is left unused.\"\n", wantSize: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Files(tt.plan)

			var data string
			for _, file := range got {
				if file.Path == tt.path {
					data = string(file.Data)
				}
			}
			if err != nil || len(got) != tt.wantSize || data != tt.want {
				t.Errorf("got %d files, %v, want %d with %s as %q", len(got), err, tt.wantSize, tt.path, tt.want)
			}
		})
	}
}

func TestNotice(t *testing.T) {
	tests := []struct {
		name  string
		given []policy.Policy
		want  string
	}{
		{name: "when Notice with a policy for review then returns no notice", given: []policy.Policy{{ID: "a"}, {ID: "b", Review: true}}, want: ""},
		{name: "when Notice with no policy for review then returns a notice", given: []policy.Policy{{ID: "a"}}, want: "no policy is marked for review, so there is no review phase: mark one with review = true in .shampoo/policies/\n"},
		{name: "when Notice with no policy then returns a notice", given: nil, want: "no policy is marked for review, so there is no review phase: mark one with review = true in .shampoo/policies/\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Notice(tt.given)

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
