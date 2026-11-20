//go:build small

package settings

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	defaults := map[string]int{"small": 60, "medium": 300, "large": 900}
	tests := []struct {
		name    string
		data    string
		want    Settings
		wantErr bool
	}{
		{name: "when Parse with no settings then returns the defaults", data: "", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with failed attempts of harness repair then returns them", data: "[state-machine.harness-repair]\nmax-failed-attempts = 1\n", want: Settings{RepairAttempts: 1, MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with a task source then returns it", data: "[orchestrator]\ntasks = \"todo.md\"\n", want: Settings{RepairAttempts: 3, Tasks: "todo.md", MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with a wait limit and a merge method then returns them", data: "[orchestrator]\nmax-wait-minutes = 0\nmerge = \"squash\"\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 0, Merge: "squash", MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with a negative wait limit then returns an error", data: "[orchestrator]\nmax-wait-minutes = -1\n", wantErr: true},
		{name: "when Parse with an unknown merge method then returns an error", data: "[orchestrator]\nmerge = \"fast-forward\"\n", wantErr: true},
		{name: "when Parse with requirements of the review then returns them as they are", data: "[state-machine.review.requirements]\nmodel = \"opus\"\nseparate-agent = true\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", Requirements: map[string]any{"model": "opus", "separate-agent": true}, MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with an unknown key of the review then returns an error", data: "[state-machine.review]\nmodel = \"opus\"\n", wantErr: true},
		{name: "when Parse with an agent of the orchestrator then returns an error", data: "[orchestrator]\nagent = \"claude -p\"\n", wantErr: true},
		{name: "when Parse with a time limit of a test size then returns it with the other defaults", data: "[test-sizes.small]\nmax-seconds = 1\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: map[string]int{"small": 1, "medium": 300, "large": 900}, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with a time limit of zero then returns an error", data: "[test-sizes.medium]\nmax-seconds = 0\n", wantErr: true},
		{name: "when Parse with a time limit of an unknown size then returns an error", data: "[test-sizes.huge]\nmax-seconds = 10\n", wantErr: true},
		{name: "when Parse with a stop of the time budget then returns it with the default budget", data: "[time-budget]\nstop = { violations = 1, runs = 1 }\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: defaults, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 1, Runs: 1}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 2, Runs: 5}}},
		{name: "when Parse with a time budget of more violations than runs then returns an error", data: "[time-budget]\nbudget = { violations = 3, runs = 2 }\n", wantErr: true},
		{name: "when Parse with a time stop of no violations then returns an error", data: "[time-budget]\nstop = { violations = 0, runs = 5 }\n", wantErr: true},
		{name: "when Parse with a stop of the workflow failures then returns it with the default budget", data: "[workflow-failure]\nstop = { violations = 1, runs = 1 }\n", want: Settings{RepairAttempts: 3, MaxWaitMinutes: 60, Merge: "merge", MaxSeconds: map[string]int{"small": 60, "medium": 300, "large": 900}, TimeBudget: Count{Violations: 2, Runs: 20}, TimeStop: Count{Violations: 2, Runs: 5}, FailureBudget: Count{Violations: 2, Runs: 20}, FailureStop: Count{Violations: 1, Runs: 1}}},
		{name: "when Parse with a stop of more workflow failures than runs then returns an error", data: "[workflow-failure]\nstop = { violations = 2, runs = 1 }\n", wantErr: true},
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
