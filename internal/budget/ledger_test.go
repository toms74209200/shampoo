//go:build small

package budget

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestParseLedger(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []Line
	}{
		{name: "when ParseLedger with lines then returns them in order", data: "{\"run\":\"a\",\"task\":\"t\",\"transition\":\"edit\",\"rule\":\"r\",\"check\":\"c\",\"outcome\":\"pass\",\"millis\":3}\n{\"run\":\"b\",\"outcome\":\"task-failure\"}\n", want: []Line{{Run: "a", Task: "t", Transition: "edit", Rule: "r", Check: "c", Outcome: rule.Outcome("pass"), Millis: 3}, {Run: "b", Outcome: rule.Outcome("task-failure")}}},
		{name: "when ParseLedger with a broken last line then returns the lines before it", data: "{\"run\":\"a\"}\n{\"run\":\"b\",\"ta", want: []Line{{Run: "a"}}},
		{name: "when ParseLedger with no data then returns no line", data: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLedger([]byte(tt.data))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestViolations(t *testing.T) {
	target := rand.Text()
	other := rand.Text()
	tests := []struct {
		name       string
		transition rule.Transition
		runs       int
		lines      []Line
		want       int
	}{
		{name: "when Violations with violations in the latest runs then returns their count", transition: rule.Transition("edit"), runs: 2, lines: []Line{{Rule: target, Transition: "edit", Outcome: rule.Outcome("task-failure")}, {Rule: target, Transition: "edit", Outcome: rule.Outcome("pass")}, {Rule: other, Transition: "edit", Outcome: rule.Outcome("task-failure")}, {Rule: target, Transition: "edit", Outcome: rule.Outcome("budget")}}, want: 1},
		{name: "when Violations with violations at another transition then returns the count at its own", transition: rule.Transition("edit"), runs: 2, lines: []Line{{Rule: target, Transition: "edit", Outcome: rule.Outcome("task-failure")}, {Rule: target, Transition: "task", Outcome: rule.Outcome("task-failure")}, {Rule: target, Transition: "task", Outcome: rule.Outcome("task-failure")}}, want: 1},
		{name: "when Violations with fewer runs than asked then returns the violations there are", transition: rule.Transition("edit"), runs: 4, lines: []Line{{Rule: target, Transition: "edit", Outcome: rule.Outcome("task-failure")}}, want: 1},
		{name: "when Violations with a workflow failure then returns zero", transition: rule.Transition("edit"), runs: 2, lines: []Line{{Rule: target, Transition: "edit", Outcome: rule.Outcome("workflow-failure")}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Violations(tt.lines, target, tt.transition, tt.runs)

			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRepairs(t *testing.T) {
	task := rand.Text()
	tests := []struct {
		name  string
		lines []Line
		want  int
	}{
		{name: "when Repairs with budgets exceeded in the task then returns their count", lines: []Line{{Task: task, Outcome: rule.Outcome("budget-exceeded")}, {Task: task, Outcome: rule.Outcome("task-failure")}, {Task: rand.Text(), Outcome: rule.Outcome("budget-exceeded")}, {Task: task, Outcome: rule.Outcome("budget-exceeded")}}, want: 2},
		{name: "when Repairs with no budget exceeded then returns zero", lines: []Line{{Task: task, Outcome: rule.Outcome("task-failure")}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Repairs(tt.lines, task)

			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	open := rand.Text()
	merged := rand.Text()
	tests := []struct {
		name   string
		lines  []Line
		window int
		want   Compaction
	}{
		{
			name: "when Compact with a merged task outside the window then returns its summary without its lines",
			lines: []Line{
				{Run: "1", Task: merged, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 5},
				{Run: "1", Task: merged, Transition: "edit", Rule: "mocks", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 5},
				{Run: "2", Task: merged, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 7},
				{Run: "3", Task: open, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 2},
				{Run: "4", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 1},
			},
			window: 1,
			want: Compaction{Kept: []Line{
				{Run: "1", Rule: "mocks", Check: "lint", Transition: "edit", Outcome: rule.Outcome("pass"), Millis: 5},
				{Run: "3", Task: open, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 2},
				{Run: "4", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 1},
			},
				Summaries: []Summary{{Task: merged, Attempts: map[string]int{"edit": 2}, Violations: map[string]int{"lint": 1}, Findings: map[string]int{}, Millis: map[string]int64{"lint": 12}}}},
		},
		{
			name: "when Compact with a merged task inside the window then returns its summary with its repairs keeping its lines without the task",
			lines: []Line{
				{Run: "1", Task: merged, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 5},
				{Run: "1", Task: merged, Transition: "edit", Rule: "mocks", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 5},
				{Run: "2", Task: merged, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 7},
				{Run: "2", Task: merged, Transition: "edit", Outcome: rule.Outcome("budget-exceeded")},
				{Run: "3", Task: open, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 2},
				{Run: "4", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 1},
			},
			window: 4,
			want: Compaction{Kept: []Line{
				{Run: "1", Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 5},
				{Run: "1", Transition: "edit", Rule: "mocks", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 5},
				{Run: "2", Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 7},
				{Run: "2", Transition: "edit", Outcome: rule.Outcome("budget-exceeded")},
				{Run: "3", Task: open, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 2},
				{Run: "4", Rule: "lint", Check: "lint", Outcome: rule.Outcome("task-failure"), Millis: 1},
			},
				Summaries: []Summary{{Task: merged, Attempts: map[string]int{"edit": 2}, Violations: map[string]int{"lint": 1}, Findings: map[string]int{}, Millis: map[string]int64{"lint": 12}, Repairs: 1}}},
		},
		{
			name: "when Compact with findings of a merged task then returns its findings by policy keeping them apart from rules in the window",
			lines: []Line{
				{Run: "1", Task: merged, Transition: "review", Policy: "functions", Outcome: rule.Outcome("finding")},
				{Run: "2", Task: merged, Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 3},
			},
			window: 1,
			want: Compaction{Kept: []Line{
				{Run: "1", Transition: "review", Policy: "functions", Outcome: rule.Outcome("finding")},
				{Run: "2", Transition: "edit", Rule: "lint", Check: "lint", Outcome: rule.Outcome("pass"), Millis: 3},
			},
				Summaries: []Summary{{Task: merged, Attempts: map[string]int{"review": 1, "edit": 1}, Violations: map[string]int{}, Findings: map[string]int{"functions": 1}, Millis: map[string]int64{"lint": 3}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compact(tt.lines, []string{open}, tt.window)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
