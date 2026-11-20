//go:build small

package budget

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestChargeRule(t *testing.T) {
	id := rand.Text()
	tests := []struct {
		name         string
		lines        []Line
		checked      rule.Checked
		transition   rule.Transition
		outcome      rule.Outcome
		exceptions   int
		failures     Limits
		wantExceeded bool
		wantQuiet    bool
		wantReport   string
	}{
		{name: "when ChargeRule with a pass then returns no charge", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("pass")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}}, transition: rule.Transition("edit"), outcome: rule.Outcome("pass"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeRule with a violation within the budget then returns no charge", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("task-failure")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}}, transition: rule.Transition("edit"), outcome: rule.Outcome("task-failure"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeRule with violations reaching the stop then returns exceeded", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("task-failure")}, {Rule: id, Transition: "edit", Outcome: rule.Outcome("budget")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}}, transition: rule.Transition("edit"), outcome: rule.Outcome("budget"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: true, wantQuiet: false, wantReport: "exceeds its budget at edit: 2 violations in the latest 5 runs"},
		{name: "when ChargeRule with a violation and a zero budget then returns exceeded", lines: []Line{{Rule: id, Transition: "task", Outcome: rule.Outcome("task-failure")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"task": {Violations: 0, Runs: 1}}, Stop: map[rule.Transition]rule.Count{"task": {Violations: 2, Runs: 5}}}, transition: rule.Transition("task"), outcome: rule.Outcome("task-failure"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: true, wantQuiet: false, wantReport: "exceeds its budget at task"},
		{name: "when ChargeRule with workflow failures reaching the stop then returns exceeded", lines: []Line{{Rule: id, Transition: "task", Outcome: rule.Outcome("workflow-failure")}, {Rule: id, Transition: "edit", Outcome: rule.Outcome("workflow-failure")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}}, transition: rule.Transition("edit"), outcome: rule.Outcome("workflow-failure"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: true, wantQuiet: false, wantReport: "exceeds its workflow-failure budget: 2 workflow failures in the latest 5 runs"},
		{name: "when ChargeRule with one workflow failure then returns no charge", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("workflow-failure")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}}, transition: rule.Transition("edit"), outcome: rule.Outcome("workflow-failure"), failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeRule with exceptions at the limit then returns no charge", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("pass")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}, Exceptions: &rule.Exceptions{File: "out/exceptions.txt", Limit: 3}}, transition: rule.Transition("edit"), outcome: rule.Outcome("pass"), exceptions: 3, failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeRule with exceptions over the limit then returns exceeded", lines: []Line{{Rule: id, Transition: "edit", Outcome: rule.Outcome("pass")}}, checked: rule.Checked{ID: id, Budget: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 20}}, Stop: map[rule.Transition]rule.Count{"edit": {Violations: 2, Runs: 5}}, Exceptions: &rule.Exceptions{File: "out/exceptions.txt", Limit: 3}}, transition: rule.Transition("edit"), outcome: rule.Outcome("pass"), exceptions: 4, failures: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: true, wantQuiet: false, wantReport: "has 4 exceptions in out/exceptions.txt, over its limit of 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChargeRule(tt.lines, tt.checked, tt.transition, tt.outcome, &tt.exceptions, tt.failures)

			if got.Exceeded != tt.wantExceeded || !strings.Contains(got.Report, tt.wantReport) || (tt.wantQuiet && got.Report != tt.wantReport) {
				t.Errorf("got %#v, want exceeded %v and a report with %q", got, tt.wantExceeded, tt.wantReport)
			}
		})
	}
}

func TestChargeTime(t *testing.T) {
	id := rand.Text()
	tests := []struct {
		name         string
		lines        []Line
		check        rule.Check
		seconds      int
		time         Limits
		wantExceeded bool
		wantQuiet    bool
		wantReport   string
	}{
		{name: "when ChargeTime with runs within the limit then returns no charge", lines: []Line{{Run: "a", Check: id, Millis: 1000}, {Run: "b", Check: id, Millis: 900}}, check: rule.Check{ID: id, Size: rule.Size("small")}, seconds: 1, time: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeTime with one run over the limit then returns no charge", lines: []Line{{Run: "a", Check: id, Millis: 1001}}, check: rule.Check{ID: id, Size: rule.Size("small")}, seconds: 1, time: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: true, wantReport: ""},
		{name: "when ChargeTime with runs over the limit reaching the stop then returns exceeded", lines: []Line{{Run: "a", Check: id, Millis: 1001}, {Run: "b", Check: id, Millis: 2000}}, check: rule.Check{ID: id, Size: rule.Size("small")}, seconds: 1, time: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: true, wantQuiet: false, wantReport: "exceeds its time budget: 2 runs over 1s of small checks in the latest 5 runs"},
		{name: "when ChargeTime with runs over the limit before the stop's window then returns a burnt budget", lines: []Line{{Run: "a", Check: id, Millis: 1001}, {Run: "b", Check: id, Millis: 2000}, {Run: "c", Check: id}, {Run: "d", Check: id}, {Run: "e", Check: id}, {Run: "f", Check: id}, {Run: "g", Check: id}}, check: rule.Check{ID: id, Size: rule.Size("medium")}, seconds: 1, time: Limits{Budget: rule.Count{Violations: 2, Runs: 20}, Stop: rule.Count{Violations: 2, Runs: 5}}, wantExceeded: false, wantQuiet: false, wantReport: "burns its time budget: 2 runs over 1s of medium checks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChargeTime(tt.lines, tt.check, tt.seconds, tt.time)

			if got.Exceeded != tt.wantExceeded || !strings.Contains(got.Report, tt.wantReport) || (tt.wantQuiet && got.Report != tt.wantReport) {
				t.Errorf("got %#v, want exceeded %v and a report with %q", got, tt.wantExceeded, tt.wantReport)
			}
		})
	}
}
