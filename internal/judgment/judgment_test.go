//go:build small

package judgment

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
)

func TestJudge(t *testing.T) {
	tests := []struct {
		name        string
		run         string
		task        string
		transition  rule.Transition
		selected    []rule.Checked
		order       []rule.Check
		lines       []budget.Line
		options     settings.Settings
		results     map[string]Result
		failing     map[string]bool
		files       map[string]string
		want        rule.Outcome
		wantErr     bool
		wantReport  string
		wantRan     []string
		wantRemoved []string
		wantPassed  bool
		wantLines   int
	}{
		{
			name:       "when Judge with a passing check then returns pass and records it",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "lint", Command: "lint.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"lint.sh": {Code: 0}},
			want:       rule.Outcome("pass"),
			wantReport: "",
			wantRan:    []string{"lint.sh"},
			wantPassed: true,
			wantLines:  1,
		},
		{
			name:       "when Judge with a check judged by its exit as a violation then returns a task failure with its output",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Statement: "Lint passes.", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "lint", Command: "lint.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"lint.sh": {Code: 1, Output: []byte("main.go:1: unused")}},
			want:       rule.Outcome("task-failure"),
			wantReport: "output of check lint:\nmain.go:1: unused",
			wantRan:    []string{"lint.sh"},
			wantPassed: false,
			wantLines:  1,
		},
		{
			name:       "when Judge with a judge command then returns its outcome and its output",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "tests", Judge: rule.ByCommand{Command: "judge.sh", Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "tests", Command: "test.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"test.sh": {Code: 0}, "judge.sh": {Code: 1, Output: []byte("coverage dropped")}},
			want:       rule.Outcome("task-failure"),
			wantReport: "judge: judge.sh exited 1\ncoverage dropped",
			wantRan:    []string{"test.sh", "judge.sh"},
			wantPassed: false,
			wantLines:  1,
		},
		{
			name:       "when Judge with a task failure and a workflow failure then returns the workflow failure",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "a", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}, {ID: "b", Check: "build", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "build", Command: "build.sh", Size: "small"}, {ID: "lint", Command: "lint.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"build.sh": {Code: 2}, "lint.sh": {Code: 1}},
			want:       rule.Outcome("workflow-failure"),
			wantReport: "check lint: lint.sh exited 1",
			wantRan:    []string{"build.sh", "lint.sh"},
			wantPassed: false,
			wantLines:  2,
		},
		{
			name:       "when Judge with a violation reaching the stop then returns exceeded and runs no later check",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "build", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 0, Runs: 1}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 1, Runs: 1}}}, {ID: "s", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "build", Command: "build.sh", Size: "small"}, {ID: "lint", Command: "lint.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"build.sh": {Code: 1}},
			want:       rule.Outcome("budget-exceeded"),
			wantReport: "repair the harness",
			wantRan:    []string{"build.sh"},
			wantPassed: false,
			wantLines:  2,
		},
		{
			name:       "when Judge with a stop reached past the repairs allowed then returns handed",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "build", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 0, Runs: 1}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 1, Runs: 1}}}},
			order:      []rule.Check{{ID: "build", Command: "build.sh", Size: "small"}},
			lines:      []budget.Line{{Task: "task-a", Outcome: rule.Outcome("budget-exceeded")}},
			options:    settings.Settings{RepairAttempts: 1, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"build.sh": {Code: 1}},
			want:       rule.Outcome("repair-exceeded"),
			wantReport: "failed repairs in task-a exceed their limit of 1",
			wantRan:    []string{"build.sh"},
			wantPassed: false,
			wantLines:  2,
		},
		{
			name:        "when Judge with more exceptions written than the limit then returns exceeded",
			run:         "run",
			task:        "task-a",
			transition:  rule.Transition("edit"),
			selected:    []rule.Checked{{ID: "r", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Exceptions: &rule.Exceptions{File: "out/exceptions.txt", Limit: 1}}},
			order:       []rule.Check{{ID: "lint", Command: "lint.sh", Size: "small"}},
			options:     settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:     map[string]Result{"lint.sh": {Code: 0}},
			files:       map[string]string{"out/exceptions.txt": "a.go:1 reason\nb.go:2 reason\n"},
			want:        rule.Outcome("budget-exceeded"),
			wantReport:  "has 2 exceptions in out/exceptions.txt, over its limit of 1",
			wantRan:     []string{"lint.sh"},
			wantRemoved: []string{"out/exceptions.txt"},
			wantPassed:  false,
			wantLines:   2,
		},
		{
			name:        "when Judge with exceptions the check did not write then returns a workflow failure",
			run:         "run",
			task:        "task-a",
			transition:  rule.Transition("edit"),
			selected:    []rule.Checked{{ID: "r", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Exceptions: &rule.Exceptions{File: "out/exceptions.txt", Limit: 3}}},
			order:       []rule.Check{{ID: "lint", Command: "lint.sh", Size: "small"}},
			options:     settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:     map[string]Result{"lint.sh": {Code: 0}},
			want:        rule.Outcome("workflow-failure"),
			wantReport:  "check lint did not write its exceptions to out/exceptions.txt",
			wantRan:     []string{"lint.sh"},
			wantRemoved: []string{"out/exceptions.txt"},
			wantPassed:  false,
			wantLines:   1,
		},
		{
			name:       "when Judge with a check that cannot run then returns an error",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "lint", Judge: rule.ByExit{Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "lint", Command: "lint.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			failing:    map[string]bool{"lint.sh": true},
			want:       rule.Outcome("workflow-failure"),
			wantErr:    true,
			wantReport: "",
			wantRan:    []string{"lint.sh"},
			wantPassed: false,
			wantLines:  0,
		},
		{
			name:       "when Judge with a judge command that cannot run then returns an error",
			run:        "run",
			task:       "task-a",
			transition: rule.Transition("edit"),
			selected:   []rule.Checked{{ID: "r", Check: "tests", Judge: rule.ByCommand{Command: "judge.sh", Exit: rule.Exit{Codes: map[int]rule.Outcome{1: "task-failure"}, Other: "workflow-failure"}}, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 5, Runs: 10}}}},
			order:      []rule.Check{{ID: "tests", Command: "test.sh", Size: "small"}},
			options:    settings.Settings{RepairAttempts: 3, MaxSeconds: map[string]int{"small": 60}, TimeBudget: rule.Count{Violations: 2, Runs: 20}, TimeStop: rule.Count{Violations: 2, Runs: 5}, FailureBudget: rule.Count{Violations: 2, Runs: 20}, FailureStop: rule.Count{Violations: 2, Runs: 5}},
			results:    map[string]Result{"test.sh": {Code: 0}},
			failing:    map[string]bool{"judge.sh": true},
			want:       rule.Outcome("workflow-failure"),
			wantErr:    true,
			wantReport: "",
			wantRan:    []string{"test.sh", "judge.sh"},
			wantPassed: false,
			wantLines:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ran, removed []string
			var appended []budget.Line
			passed := false
			world := World{
				Run: func(command string) (Result, error) {
					ran = append(ran, command)
					if tt.failing[command] {
						return Result{}, errors.New(command)
					}
					return tt.results[command], nil
				},
				Remove: func(path string) error {
					removed = append(removed, path)
					return nil
				},
				Read: func(path string) ([]byte, error) {
					content, ok := tt.files[path]
					if !ok {
						return nil, errors.New(path)
					}
					return []byte(content), nil
				},
				Append: func(line budget.Line) error {
					appended = append(appended, line)
					return nil
				},
				Pass: func() error {
					passed = true
					return nil
				},
			}
			session := Session{Run: tt.run, Task: tt.task, Transition: tt.transition, Selected: tt.selected, Order: tt.order, Lines: tt.lines, Options: tt.options}

			got, report, err := Judge(session, world)

			if got != tt.want || (err != nil) != tt.wantErr || !strings.Contains(report, tt.wantReport) {
				t.Errorf("got %q, %v, report %q; want %q, error %v, report with %q", got, err, report, tt.want, tt.wantErr, tt.wantReport)
			}
			if !reflect.DeepEqual(ran, tt.wantRan) || !reflect.DeepEqual(removed, tt.wantRemoved) || passed != tt.wantPassed || len(appended) != tt.wantLines {
				t.Errorf("ran %v, removed %v, passed %v, appended %d lines; want %v, %v, %v, %d", ran, removed, passed, len(appended), tt.wantRan, tt.wantRemoved, tt.wantPassed, tt.wantLines)
			}
		})
	}
}

func TestParseAttempt(t *testing.T) {
	tests := []struct {
		name string
		data string
		tree string
		want Earlier
	}{
		{name: "when ParseAttempt with an attempt on the same tree then returns it cached", data: `{"tree": "t1", "verdict": "task-failure", "report": "r"}`, tree: "t1", want: Cached{Verdict: rule.Outcome("task-failure"), Report: "r"}},
		{name: "when ParseAttempt with an attempt on another tree then returns fresh", data: `{"tree": "t0", "verdict": "pass"}`, tree: "t1", want: Fresh{}},
		{name: "when ParseAttempt with no attempt then returns fresh", data: "", tree: "t1", want: Fresh{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAttempt([]byte(tt.data), tt.tree)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
