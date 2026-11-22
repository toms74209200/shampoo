//go:build small

package budget

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestDiverged(t *testing.T) {
	task, other := rand.Text(), rand.Text()
	tests := []struct {
		name  string
		lines []Line
		limit Reviews
		want  bool
	}{
		{name: "when Diverged with one review then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 0, Max: 9}, want: false},
		{name: "when Diverged with findings that drop then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 0, Max: 9}, want: false},
		{name: "when Diverged with findings that do not drop beyond the attempts then returns true", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 0, Max: 9}, want: true},
		{name: "when Diverged with findings that do not drop within the attempts then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 1, Max: 9}, want: false},
		{name: "when Diverged with findings that dropped after not dropping then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "c", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 0, Max: 9}, want: false},
		{name: "when Diverged with findings that drop beyond the reviews then returns true", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 9, Max: 1}, want: true},
		{name: "when Diverged with findings that drop within the reviews then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 9, Max: 2}, want: false},
		{name: "when Diverged with a review without findings beyond the reviews then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("pass")}}, limit: Reviews{Attempts: 9, Max: 1}, want: false},
		{name: "when Diverged with a review without findings then returns false", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("pass")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("pass")}}, limit: Reviews{Attempts: 0, Max: 9}, want: false},
		{name: "when Diverged with reviews of another task then returns false", lines: []Line{{Run: "a", Task: other, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: other, Transition: "review", Outcome: rule.Outcome("finding")}}, limit: Reviews{Attempts: 0, Max: 9}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Diverged(tt.lines, task, tt.limit)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewed(t *testing.T) {
	run := rand.Text()
	tests := []struct {
		name     string
		task     string
		ledger   string
		policies []string
		limit    Reviews
		want     []Line
	}{
		{name: "when Reviewed with findings then returns a pass line and a line per finding", task: "task-a", ledger: "", policies: []string{"functions", "tests"}, limit: Reviews{Attempts: 0, Max: 9}, want: []Line{{Run: run, Task: "task-a", Transition: "review", Outcome: rule.Outcome("pass")}, {Run: run, Task: "task-a", Transition: "review", Policy: "functions", Outcome: rule.Outcome("finding")}, {Run: run, Task: "task-a", Transition: "review", Policy: "tests", Outcome: rule.Outcome("finding")}}},
		{name: "when Reviewed with findings that do not drop beyond the attempts then returns an entry into repair", task: "task-a", ledger: "{\"run\":\"a\",\"task\":\"task-a\",\"transition\":\"review\",\"outcome\":\"finding\"}\n", policies: []string{"functions"}, limit: Reviews{Attempts: 0, Max: 9}, want: []Line{{Run: run, Task: "task-a", Transition: "review", Outcome: rule.Outcome("pass")}, {Run: run, Task: "task-a", Transition: "review", Policy: "functions", Outcome: rule.Outcome("finding")}, {Run: run, Task: "task-a", Transition: "review", Outcome: rule.Outcome("budget-exceeded")}}},
		{name: "when Reviewed with no findings then returns the pass line only", task: "task-a", ledger: "", policies: nil, limit: Reviews{Attempts: 0, Max: 9}, want: []Line{{Run: run, Task: "task-a", Transition: "review", Outcome: rule.Outcome("pass")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reviewed([]byte(tt.ledger), run, tt.task, tt.policies, tt.limit)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestConverge(t *testing.T) {
	task := rand.Text()
	tests := []struct {
		name   string
		lines  []Line
		review Reviews
		repair int
		want   Convergence
	}{
		{name: "when Converge with findings that drop then returns converging", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}}, review: Reviews{Attempts: 0, Max: 9}, repair: 1, want: Converging{}},
		{name: "when Converge with findings that do not drop within the repair attempts then returns diverging", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("budget-exceeded")}}, review: Reviews{Attempts: 0, Max: 9}, repair: 1, want: Diverging{}},
		{name: "when Converge with findings that do not drop beyond the repair attempts then returns exhausted", lines: []Line{{Run: "a", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("finding")}, {Run: "b", Task: task, Transition: "review", Outcome: rule.Outcome("budget-exceeded")}}, review: Reviews{Attempts: 0, Max: 9}, repair: 0, want: Exhausted{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Converge(tt.lines, task, tt.review, tt.repair)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
