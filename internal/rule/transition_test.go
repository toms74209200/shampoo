//go:build small

package rule

import "testing"

func TestParseTransition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Transition
		wantErr bool
	}{
		{name: "when ParseTransition with edit then returns the edit transition", input: "edit", want: Transition("edit")},
		{name: "when ParseTransition with task then returns the task transition", input: "task", want: Transition("task")},
		{name: "when ParseTransition with pr then returns the pr transition", input: "pr", want: Transition("pr")},
		{name: "when ParseTransition with milestone then returns the milestone transition", input: "milestone", want: Transition("milestone")},
		{name: "when ParseTransition with an unknown name then returns an error", input: "merge", wantErr: true},
		{name: "when ParseTransition with an empty name then returns an error", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransition(tt.input)

			if got != tt.want || (err != nil) != tt.wantErr {
				t.Errorf("got %q, %v, want %q and an error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
