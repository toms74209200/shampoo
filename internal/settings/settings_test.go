//go:build small

package settings

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Settings
		wantErr bool
	}{
		{name: "when Parse with no settings then returns the defaults", data: "", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge"}},
		{name: "when Parse with failed attempts of harness repair then returns them", data: "[state-machine.harness-repair]\nmax-failed-attempts = 1\n", want: Settings{RepairAttempts: 1, MaxWaitMinutes: 60, Merge: "merge"}},
		{name: "when Parse with a task source then returns it", data: "[orchestrator]\ntasks = \"todo.md\"\n", want: Settings{RepairAttempts: 3, Tasks: "todo.md", MaxWaitMinutes: 60, Merge: "merge"}},
		{name: "when Parse with a wait limit and a merge method then returns them", data: "[orchestrator]\nmax-wait-minutes = 0\nmerge = \"squash\"\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 0, Merge: "squash"}},
		{name: "when Parse with a negative wait limit then returns an error", data: "[orchestrator]\nmax-wait-minutes = -1\n", wantErr: true},
		{name: "when Parse with an unknown merge method then returns an error", data: "[orchestrator]\nmerge = \"fast-forward\"\n", wantErr: true},
		{name: "when Parse with requirements of the review then returns them as they are", data: "[state-machine.review.requirements]\nmodel = \"opus\"\nseparate-agent = true\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", Requirements: map[string]any{"model": "opus", "separate-agent": true}}},
		{name: "when Parse with an unknown key of the review then returns an error", data: "[state-machine.review]\nmodel = \"opus\"\n", wantErr: true},
		{name: "when Parse with an agent of the orchestrator then returns an error", data: "[orchestrator]\nagent = \"claude -p\"\n", wantErr: true},
		{name: "when Parse with invalid TOML then returns an error", data: "[state-machine", wantErr: true},
		{name: "when Parse with a state not settable yet then returns an error", data: "[state-machine.editing]\nmax-failed-attempts = 3\n", wantErr: true},
		{name: "when Parse with negative failed attempts then returns an error", data: "[state-machine.harness-repair]\nmax-failed-attempts = -1\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
