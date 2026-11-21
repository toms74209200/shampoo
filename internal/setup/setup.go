package setup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/BurntSushi/toml"

	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/rule"
)

var ErrInvalid = errors.New("invalid setup")

type Check struct {
	Rule       string
	Command    string
	Size       rule.Size
	Paths      []string
	Exceptions string
}

type Plan struct {
	Tasks  string
	Checks []Check
}

type File struct {
	Path string
	Data []byte
}

type recommended struct {
	statement   string
	transitions []string
}

var rules = map[string]recommended{
	"lint":           {statement: "The code passes the project's linter.", transitions: []string{"edit", "task", "pr"}},
	"small-tests":    {statement: "Every small test passes.", transitions: []string{"edit", "task", "pr"}},
	"coverage":       {statement: "Small-test coverage is at least the project's threshold.", transitions: []string{"task", "pr"}},
	"commit-message": {statement: "Every commit message follows the project's format.", transitions: []string{"pr"}},
}

var policies = map[string]string{
	"tests": `Code and tests are two sides of one thing. Tests are as many small tests
as possible.
- Code with side effects is tested by medium or larger tests, written knowing their cost
- DAMP over DRY
- Arrange-Act-Assert`,
	"functions":     "A function adds something its callers would otherwise write: no wrapper used in one place, and no function that only calls one level down.",
	"comments":      "A comment says what the code cannot: no comment that restates the code.",
	"errors":        "Every error is handled or returned: none is ignored or only logged.",
	"scope":         "A change does the task and nothing else: no refactoring or reformatting on the way.",
	"reuse":         "What the project or its dependencies already do is used, not written again.",
	"assertions":    "Every test asserts what the code must do: none asserts nothing or restates the implementation.",
	"overfitting":   "Code solves the problem, not the test: no hard-coded value or branch for a test's case.",
	"dead-code":     "No code, argument or setting is left unused.",
	"pull-requests": "A pull request is one task, kept small: what can merge on its own is a pull request of its own.",
}

func Parse(data []byte) (Plan, error) {
	var input struct {
		Tasks  string `json:"tasks"`
		Checks map[string]struct {
			Command    string   `json:"command"`
			Size       string   `json:"size"`
			Paths      []string `json:"paths"`
			Exceptions string   `json:"exceptions"`
		} `json:"checks"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Plan{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	plan := Plan{Tasks: input.Tasks}
	for id, check := range input.Checks {
		if _, ok := rules[id]; !ok {
			return Plan{}, fmt.Errorf("%w: %q is not a recommended rule", ErrInvalid, id)
		}
		if check.Command == "" {
			return Plan{}, fmt.Errorf("%w: the check of %s has no command", ErrInvalid, id)
		}
		if check.Exceptions != "" && id != "lint" {
			return Plan{}, fmt.Errorf("%w: only lint writes exceptions, not %s", ErrInvalid, id)
		}
		size := check.Size
		if size == "" {
			size = "small"
		}
		if !slices.Contains([]rule.Size{rule.Small, rule.Medium, rule.Large}, rule.Size(size)) {
			return Plan{}, fmt.Errorf("%w: %s: size %q", ErrInvalid, id, size)
		}
		plan.Checks = append(plan.Checks, Check{Rule: id, Command: check.Command, Size: rule.Size(size), Paths: check.Paths, Exceptions: check.Exceptions})
	}
	sort.Slice(plan.Checks, func(i, j int) bool { return plan.Checks[i].Rule < plan.Checks[j].Rule })
	return plan, nil
}

func Files(plan Plan) ([]File, error) {
	type exceptions struct {
		File string `toml:"file"`
	}
	type ruleFile struct {
		Rule        string         `toml:"rule"`
		Check       string         `toml:"check"`
		Transitions []string       `toml:"transitions"`
		Paths       []string       `toml:"paths,omitempty"`
		Judge       map[string]any `toml:"judge"`
		Exceptions  *exceptions    `toml:"exceptions,omitempty"`
	}
	type checkFile struct {
		Command string `toml:"command"`
		Size    string `toml:"size"`
	}
	type policyFile struct {
		Review bool   `toml:"review"`
		Text   string `toml:"text"`
	}
	settings := map[string]any{"orchestrator": map[string]string{"tasks": plan.Tasks}}
	documents := []struct {
		path  string
		value any
	}{{path: ".shampoo/settings.toml", value: settings}}
	for _, check := range plan.Checks {
		recommended := rules[check.Rule]
		file := ruleFile{Rule: recommended.statement, Check: check.Rule, Transitions: recommended.transitions, Paths: check.Paths, Judge: map[string]any{"exit": map[string]string{"1": "task-failure", "other": "workflow-failure"}}}
		if check.Exceptions != "" {
			file.Exceptions = &exceptions{File: check.Exceptions}
		}
		documents = append(documents,
			struct {
				path  string
				value any
			}{path: ".shampoo/checks/" + check.Rule + ".toml", value: checkFile{Command: check.Command, Size: string(check.Size)}},
			struct {
				path  string
				value any
			}{path: ".shampoo/rules/" + check.Rule + ".toml", value: file})
	}
	ids := make([]string, 0, len(policies))
	for id := range policies {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		documents = append(documents, struct {
			path  string
			value any
		}{path: ".shampoo/policies/" + id + ".toml", value: policyFile{Review: true, Text: policies[id]}})
	}
	var files []File
	for _, document := range documents {
		var data bytes.Buffer
		if err := toml.NewEncoder(&data).Encode(document.value); err != nil {
			return nil, err
		}
		files = append(files, File{Path: document.path, Data: data.Bytes()})
	}
	return files, nil
}

func Notice(policies []policy.Policy) string {
	if slices.ContainsFunc(policies, func(p policy.Policy) bool { return p.Review }) {
		return ""
	}
	return "no policy is marked for review, so there is no review phase: mark one with review = true in .shampoo/policies/\n"
}
