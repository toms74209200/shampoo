//go:build small

package budget

import (
	"crypto/rand"
	"math"
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

func TestBurnRate(t *testing.T) {
	target := rand.Text()
	other := rand.Text()
	tests := []struct {
		name  string
		runs  int
		lines []Line
		limit float64
		want  float64
	}{
		{name: "when BurnRate with violations in the latest runs then returns their share over the limit", runs: 2, lines: []Line{{Rule: target, Outcome: rule.Outcome("task-failure")}, {Rule: target, Outcome: rule.Outcome("pass")}, {Rule: other, Outcome: rule.Outcome("task-failure")}, {Rule: target, Outcome: rule.Outcome("budget")}}, limit: 0.5, want: 1},
		{name: "when BurnRate with fewer runs than its window then returns the missing runs as non-violations", runs: 4, lines: []Line{{Rule: target, Outcome: rule.Outcome("task-failure")}}, limit: 0.5, want: 0.5},
		{name: "when BurnRate with a workflow failure then returns no burn", runs: 2, lines: []Line{{Rule: target, Outcome: rule.Outcome("workflow-failure")}}, limit: 0.1, want: 0},
		{name: "when BurnRate with a violation against a zero limit then returns infinity", runs: 2, lines: []Line{{Rule: target, Outcome: rule.Outcome("task-failure")}}, limit: 0, want: math.Inf(1)},
		{name: "when BurnRate with no violation against a zero limit then returns no burn", runs: 2, lines: []Line{{Rule: target, Outcome: rule.Outcome("pass")}}, limit: 0, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BurnRate(tt.lines, target, tt.runs, tt.limit)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
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
				Summaries: []Summary{{Task: merged, Attempts: map[string]int{"edit": 2}, Violations: map[string]int{"lint": 1}, Millis: map[string]int64{"lint": 12}}}},
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
				Summaries: []Summary{{Task: merged, Attempts: map[string]int{"edit": 2}, Violations: map[string]int{"lint": 1}, Millis: map[string]int64{"lint": 12}, Repairs: 1}}},
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
