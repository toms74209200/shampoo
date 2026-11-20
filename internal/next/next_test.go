//go:build small

package next

import (
	"crypto/rand"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestDecide(t *testing.T) {
	source := rand.Text()
	branch := rand.Text()
	tests := []struct {
		name  string
		state State
		want  Action
	}{
		{name: "when Decide with the default branch then returns taking a task from the source", state: State{Source: source, Branch: branch, Default: true, Verdict: rule.Outcome("pass")}, want: Action{Action: "take-task", Source: source}},
		{name: "when Decide with a passed task guard then returns opening a pull request", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("pass")}, want: Action{Action: "open-pr", Branch: branch}},
		{name: "when Decide with a budget exceeded then returns repairing the harness", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("budget-exceeded")}, want: Action{Action: "repair-harness", Branch: branch}},
		{name: "when Decide with repairs beyond their limit then returns handing the task to a human", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("repair-exceeded")}, want: Action{Action: "hand-to-human", Branch: branch}},
		{name: "when Decide with a task failure then returns going on with the task", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("task-failure")}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with no judgment of the current tree then returns going on with the task", state: State{Source: source, Branch: branch}, want: Action{Action: "continue", Branch: branch}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.state)

			if got != tt.want {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
