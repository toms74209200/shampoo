//go:build small

package rule

import "testing"

func TestOutcome(t *testing.T) {
	tests := []struct {
		name string
		exit Exit
		code int
		want Outcome
	}{
		{name: "when Outcome with exit code zero then returns pass", exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure")}, other: Outcome("workflow-failure")}, code: 0, want: Outcome("pass")},
		{name: "when Outcome with a mapped exit code then returns its outcome", exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure")}, other: Outcome("workflow-failure")}, code: 1, want: Outcome("task-failure")},
		{name: "when Outcome with an unmapped exit code then returns the other outcome", exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure")}, other: Outcome("workflow-failure")}, code: 2, want: Outcome("workflow-failure")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.exit.Outcome(tt.code)

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
