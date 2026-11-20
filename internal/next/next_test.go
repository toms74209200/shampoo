//go:build small

package next

import (
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	"github.com/toms74209200/shampoo/internal/review"
	"github.com/toms74209200/shampoo/internal/rule"
)

func TestDecide(t *testing.T) {
	source := rand.Text()
	branch := rand.Text()
	first, second := rand.Text(), rand.Text()
	elsewhere := rand.Text()
	url, older, other := rand.Text(), rand.Text(), rand.Text()
	start := time.Now()
	pr := func(state string, pushed bool, expected []string, runs ...Run) *PullRequest {
		return &PullRequest{State: state, Pushed: pushed, Commits: []string{first, second}, Expected: expected, Runs: runs}
	}
	tests := []struct {
		name  string
		state State
		want  Action
	}{
		{name: "when Decide with the default branch then returns taking a task from the source", state: State{Source: source, Branch: branch, Default: true, Verdict: rule.Outcome("pass")}, want: Action{Action: "take-task", Source: source}},
		{name: "when Decide with the default branch and open questions within the limit then returns taking a task", state: State{Source: source, Default: true, Questions: []string{url}, MaxOpen: 1}, want: Action{Action: "take-task", Source: source}},
		{name: "when Decide with the default branch and more open questions than the limit then returns waiting for a human with them", state: State{Source: source, Default: true, Questions: []string{url, other}, MaxOpen: 1}, want: Action{Action: "wait-for-human", Questions: []string{url, other}}},
		{name: "when Decide with a task branch and more open questions than the limit then returns going on with the task", state: State{Branch: branch, Questions: []string{url, other}, MaxOpen: 1}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with a passed task guard then returns opening a pull request", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("pass")}, want: Action{Action: "open-pr", Branch: branch}},
		{name: "when Decide with a budget exceeded then returns repairing the harness", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("budget-exceeded")}, want: Action{Action: "repair-harness", Branch: branch}},
		{name: "when Decide with repairs beyond their limit then returns handing the task to a human", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("repair-exceeded")}, want: Action{Action: "hand-to-human", Branch: branch}},
		{name: "when Decide with a task failure then returns going on with the task", state: State{Source: source, Branch: branch, Verdict: rule.Outcome("task-failure")}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with no judgment of the current tree then returns going on with the task", state: State{Source: source, Branch: branch}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with a passed task guard and policies for review then returns reviewing with them and the requirements", state: State{Branch: branch, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}, want: Action{Action: "review", Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}},
		{name: "when Decide with a reviewed task then returns opening a pull request with the findings", state: State{Branch: branch, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Reviewed: true, Findings: []review.Finding{{Policy: "functions", Comment: url}}}, want: Action{Action: "open-pr", Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}},
		{name: "when Decide with a passed guard on commits the pull request lacks then returns opening a pull request", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", false, nil)}, want: Action{Action: "open-pr", Branch: branch}},
		{name: "when Decide with a task failure on an open pull request then returns going on with the task", state: State{Branch: branch, Verdict: rule.Outcome("task-failure"), PullRequest: pr("OPEN", true, nil)}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with a merged pull request then returns merged", state: State{Branch: branch, PullRequest: pr("MERGED", true, nil)}, want: Action{Action: "merged", Branch: branch}},
		{name: "when Decide with a closed pull request then returns handing the task to a human", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("CLOSED", true, nil)}, want: Action{Action: "hand-to-human", Branch: branch}},
		{name: "when Decide with no run of an expected workflow then returns waiting", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"})}, want: Action{Action: "wait", Branch: branch}},
		{name: "when Decide with an expected workflow running then returns waiting", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "in_progress", URL: url, Created: start})}, want: Action{Action: "wait", Branch: branch}},
		{name: "when Decide with an expected workflow failed then returns fixing it with its run", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "commit"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "failure", URL: url, Created: start})}, want: Action{Action: "fix-pr", Branch: branch, Failures: []Failure{{Workflow: "go", URL: url}}}},
		{name: "when Decide with every expected workflow passed then returns merging", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "docs", "lint"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "docs", Commit: second, Status: "completed", Conclusion: "skipped", Created: start}, Run{Workflow: "lint", Commit: second, Status: "completed", Conclusion: "neutral", Created: start})}, want: Action{Action: "merge", Branch: branch}},
		{name: "when Decide with a run on an earlier commit then returns merging", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start})}, want: Action{Action: "merge", Branch: branch}},
		{name: "when Decide with a later run that did not execute then returns merging by the earlier run", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "action_required", Created: start.Add(time.Minute)})}, want: Action{Action: "merge", Branch: branch}},
		{name: "when Decide with a failed run before a passed one then returns merging", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", URL: other, Created: start.Add(time.Minute)}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "failure", URL: older, Created: start})}, want: Action{Action: "merge", Branch: branch}},
		{name: "when Decide with a run on a commit outside the pull request then returns waiting", state: State{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: elsewhere, Status: "completed", Conclusion: "success", Created: start})}, want: Action{Action: "wait", Branch: branch}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.state)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
