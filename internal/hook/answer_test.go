//go:build small

package hook

import (
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestGuard(t *testing.T) {
	tests := []struct {
		name    string
		outcome rule.Outcome
		report  string
		want    Answer
	}{
		{name: "when Guard with a pass then returns the code alone", outcome: rule.Outcome("pass"), report: "", want: Answer{Code: 0, Outcome: rule.Outcome("pass")}},
		{name: "when Guard with a task failure then returns sending the agent back in every agent's fields", outcome: rule.Outcome("task-failure"), report: "found TODO", want: Answer{Code: 1, Outcome: rule.Outcome("task-failure"), Report: "found TODO", Decision: "block", Reason: "found TODO", FollowupMessage: "found TODO", AdditionalContext: "found TODO"}},
		{name: "when Guard with a stop to harness repair then returns sending the agent back", outcome: rule.Outcome("budget-exceeded"), report: "repair the harness", want: Answer{Code: 3, Outcome: rule.Outcome("budget-exceeded"), Report: "repair the harness", Decision: "block", Reason: "repair the harness", FollowupMessage: "repair the harness", AdditionalContext: "repair the harness"}},
		{name: "when Guard with a workflow failure then returns a message for the human", outcome: rule.Outcome("workflow-failure"), report: "check crashed", want: Answer{Code: 2, Outcome: rule.Outcome("workflow-failure"), Report: "check crashed", SystemMessage: "check crashed"}},
		{name: "when Guard with repairs beyond their limit then returns a message for the human", outcome: rule.Outcome("repair-exceeded"), report: "hand the task to a human", want: Answer{Code: 4, Outcome: rule.Outcome("repair-exceeded"), Report: "hand the task to a human", SystemMessage: "hand the task to a human"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Guard(tt.outcome, tt.report)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPolicies(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantEvent string
	}{
		{name: "when Policies with a text then returns it for every agent's session start", text: "# Policies", wantEvent: "SessionStart"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Policies(tt.text)

			if got.HookSpecificOutput.HookEventName != tt.wantEvent || got.HookSpecificOutput.AdditionalContext != tt.text || got.AdditionalContext != tt.text {
				t.Errorf("got %#v, want %q for %s", got, tt.text, tt.wantEvent)
			}
		})
	}
}
