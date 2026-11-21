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

func Hooks(data []byte) ([]byte, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		data = []byte("{}\n")
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("%w: .claude/settings.json is not JSON", ErrInvalid)
	}
	type span struct{ start, end int }
	members := func(at span) (map[string]span, error) {
		decoder := json.NewDecoder(bytes.NewReader(data[at.start:at.end]))
		if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
			return nil, fmt.Errorf("%w: .claude/settings.json: an object is expected", ErrInvalid)
		}
		found := map[string]span{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			var raw json.RawMessage
			if err := decoder.Decode(&raw); err != nil {
				return nil, err
			}
			end := at.start + int(decoder.InputOffset())
			found[key.(string)] = span{start: end - len(raw), end: end}
		}
		return found, nil
	}
	insert := func(at span, empty bool, text []byte) []byte {
		closing := at.end - 1
		line := bytes.LastIndexByte(data[:closing], '\n') + 1
		indent := data[line : line+len(data[line:closing])-len(bytes.TrimLeft(data[line:closing], " \t"))]
		if len(bytes.TrimSpace(data[line:closing])) > 0 {
			indent = nil
		}
		point := len(bytes.TrimRight(data[:closing], " \t\r\n"))
		var added bytes.Buffer
		if !empty {
			added.WriteByte(',')
		}
		added.WriteString("\n" + string(indent) + "  ")
		added.Write(bytes.ReplaceAll(text, []byte("\n"), []byte("\n"+string(indent)+"  ")))
		if !bytes.Contains(data[point:closing], []byte("\n")) {
			added.WriteString("\n" + string(indent))
		}
		return slices.Concat(data[:point], added.Bytes(), data[point:])
	}
	guard := func(transition string) string {
		return fmt.Sprintf("sh -c 'shampoo guard %s; case $? in 0) exit 0 ;; %d|%d) exit 2 ;; *) exit 1 ;; esac'", transition, rule.ExitCodes[rule.TaskFailure], rule.ExitCodes[rule.Exceeded])
	}
	wanted := []struct {
		event   string
		matcher string
		command string
	}{
		{event: "PostToolUse", matcher: "*", command: guard("edit")},
		{event: "Stop", command: guard("task")},
		{event: "SessionStart", matcher: "startup|clear|compact", command: "shampoo policies"},
	}
	for _, want := range wanted {
		entry := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": want.command, "timeout": 600}}}
		if want.matcher != "" {
			entry["matcher"] = want.matcher
		}
		top, err := members(span{start: 0, end: len(bytes.TrimRight(data, " \t\r\n"))})
		if err != nil {
			return nil, err
		}
		hooks, ok := top["hooks"]
		if !ok {
			text, err := json.MarshalIndent(map[string]any{want.event: []any{entry}}, "", "  ")
			if err != nil {
				return nil, err
			}
			data = insert(span{start: 0, end: len(bytes.TrimRight(data, " \t\r\n"))}, len(top) == 0, append([]byte(`"hooks": `), text...))
			continue
		}
		events, err := members(hooks)
		if err != nil {
			return nil, err
		}
		event, ok := events[want.event]
		if !ok {
			text, err := json.MarshalIndent([]any{entry}, "", "  ")
			if err != nil {
				return nil, err
			}
			data = insert(hooks, len(events) == 0, append([]byte(`"`+want.event+`": `), text...))
			continue
		}
		var present []struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		}
		if err := json.Unmarshal(data[event.start:event.end], &present); err != nil {
			return nil, fmt.Errorf("%w: .claude/settings.json: hooks of %s: %v", ErrInvalid, want.event, err)
		}
		if slices.ContainsFunc(present, func(p struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		}) bool {
			return slices.ContainsFunc(p.Hooks, func(h struct {
				Command string `json:"command"`
			}) bool {
				return h.Command == want.command
			})
		}) {
			continue
		}
		text, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return nil, err
		}
		data = insert(event, len(present) == 0, text)
	}
	return data, nil
}

func Notice(policies []policy.Policy) string {
	if slices.ContainsFunc(policies, func(p policy.Policy) bool { return p.Review }) {
		return ""
	}
	return "no policy is marked for review, so there is no review phase: mark one with review = true in .shampoo/policies/\n"
}
