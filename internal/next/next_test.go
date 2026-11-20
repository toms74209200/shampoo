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

func TestParse(t *testing.T) {
	source := rand.Text()
	branch := rand.Text()
	tree := rand.Text()
	first, second := rand.Text(), rand.Text()
	elsewhere := rand.Text()
	url, older, other := rand.Text(), rand.Text(), rand.Text()
	start := time.Now()
	pr := func(state string, pushed bool, expected []string, runs ...Run) *PullRequest {
		return &PullRequest{State: state, Pushed: pushed, Commits: []string{first, second}, Latest: start, Expected: expected, Runs: runs}
	}
	tests := []struct {
		name        string
		observation Observation
		want        State
	}{
		{name: "when Parse with the default branch then returns idle with the source", observation: Observation{Source: source, Branch: branch, Default: true, Verdict: rule.Outcome("pass")}, want: Idle{Source: source}},
		{name: "when Parse with the default branch and open questions within the limit then returns idle", observation: Observation{Source: source, Default: true, Questions: []string{url}, MaxOpen: 1}, want: Idle{Source: source}},
		{name: "when Parse with the default branch and more open questions than the limit then returns blocked by them", observation: Observation{Source: source, Default: true, Questions: []string{url, other}, MaxOpen: 1}, want: Blocked{Questions: []string{url, other}}},
		{name: "when Parse with a task branch and more open questions than the limit then returns editing", observation: Observation{Branch: branch, Questions: []string{url, other}, MaxOpen: 1}, want: Editing{Branch: branch}},
		{name: "when Parse with a passed task guard then returns ready", observation: Observation{Source: source, Branch: branch, Verdict: rule.Outcome("pass")}, want: Ready{Branch: branch}},
		{name: "when Parse with a budget exceeded then returns repairing", observation: Observation{Source: source, Branch: branch, Verdict: rule.Outcome("budget-exceeded")}, want: Repairing{Branch: branch}},
		{name: "when Parse with repairs beyond their limit then returns handed", observation: Observation{Source: source, Branch: branch, Verdict: rule.Outcome("repair-exceeded")}, want: Handed{Branch: branch}},
		{name: "when Parse with a task failure then returns editing", observation: Observation{Source: source, Branch: branch, Verdict: rule.Outcome("task-failure")}, want: Editing{Branch: branch}},
		{name: "when Parse with no judgment of the current tree then returns editing", observation: Observation{Source: source, Branch: branch}, want: Editing{Branch: branch}},
		{name: "when Parse with a passed task guard and policies for review then returns done with them and the requirements", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}, want: Done{Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}},
		{name: "when Parse with a review of the task on its tree then returns ready with the findings", observation: Observation{Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Record: &Record{Branch: branch, Tree: tree, Findings: []review.Finding{{Policy: "functions", Comment: url}}}}, want: Ready{Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}},
		{name: "when Parse with a review of the task on an earlier tree then returns ready without findings", observation: Observation{Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Record: &Record{Branch: branch, Tree: other, Findings: []review.Finding{{Policy: "functions", Comment: url}}}}, want: Ready{Branch: branch}},
		{name: "when Parse with a review of another task then returns done", observation: Observation{Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []string{"functions"}, Record: &Record{Branch: other, Tree: tree}}, want: Done{Branch: branch, Policies: []string{"functions"}}},
		{name: "when Parse with a passed guard on commits the pull request lacks then returns ready", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", false, nil)}, want: Ready{Branch: branch}},
		{name: "when Parse with a task failure on an open pull request then returns editing", observation: Observation{Branch: branch, Verdict: rule.Outcome("task-failure"), PullRequest: pr("OPEN", true, nil)}, want: Editing{Branch: branch}},
		{name: "when Parse with a merged pull request then returns merged", observation: Observation{Branch: branch, PullRequest: pr("MERGED", true, nil)}, want: Merged{Branch: branch}},
		{name: "when Parse with a closed pull request then returns handed", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("CLOSED", true, nil)}, want: Handed{Branch: branch}},
		{name: "when Parse with no run of an expected workflow then returns waiting", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}), Now: start.Add(59 * time.Minute), MaxWait: time.Hour}, want: Waiting{Branch: branch}},
		{name: "when Parse with no run of an expected workflow for the longest wait then returns stalled since the latest commit", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}), Now: start.Add(time.Hour), MaxWait: time.Hour}, want: Stalled{Branch: branch, Since: start, MaxWait: time.Hour}},
		{name: "when Parse with an expected workflow running then returns waiting", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "in_progress", URL: url, Created: start}), Now: start, MaxWait: time.Hour}, want: Waiting{Branch: branch}},
		{name: "when Parse with an expected workflow failed then returns failing with its run", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "commit"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "failure", URL: url, Created: start})}, want: Failing{Branch: branch, Failures: []Failure{{Workflow: "go", URL: url}}}},
		{name: "when Parse with every expected workflow passed then returns passed", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "docs", "lint"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "docs", Commit: second, Status: "completed", Conclusion: "skipped", Created: start}, Run{Workflow: "lint", Commit: second, Status: "completed", Conclusion: "neutral", Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a run on an earlier commit then returns passed", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a later run that did not execute then returns passed by the earlier run", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "action_required", Created: start.Add(time.Minute)})}, want: Passed{Branch: branch}},
		{name: "when Parse with a failed run before a passed one then returns passed", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", URL: other, Created: start.Add(time.Minute)}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "failure", URL: older, Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a run on a commit outside the pull request then returns waiting", observation: Observation{Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: elsewhere, Status: "completed", Conclusion: "success", Created: start}), Now: start, MaxWait: time.Hour}, want: Waiting{Branch: branch}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.observation)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecide(t *testing.T) {
	source := rand.Text()
	branch := rand.Text()
	url := rand.Text()
	tests := []struct {
		name  string
		state State
		want  Action
	}{
		{name: "when Decide with idle then returns taking a task from the source", state: Idle{Source: source}, want: Action{Action: "take-task", Source: source}},
		{name: "when Decide with blocked then returns waiting for a human with the questions", state: Blocked{Questions: []string{url}}, want: Action{Action: "wait-for-human", Questions: []string{url}}},
		{name: "when Decide with editing then returns going on with the task", state: Editing{Branch: branch}, want: Action{Action: "continue", Branch: branch}},
		{name: "when Decide with repairing then returns repairing the harness", state: Repairing{Branch: branch}, want: Action{Action: "repair-harness", Branch: branch}},
		{name: "when Decide with handed then returns handing the task to a human", state: Handed{Branch: branch}, want: Action{Action: "hand-to-human", Branch: branch}},
		{name: "when Decide with done then returns reviewing with the policies and the requirements", state: Done{Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}, want: Action{Action: "review", Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}},
		{name: "when Decide with ready then returns opening a pull request with the findings", state: Ready{Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}, want: Action{Action: "open-pr", Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}},
		{name: "when Decide with failing then returns fixing the pull request", state: Failing{Branch: branch, Failures: []Failure{{Workflow: "go", URL: url}}}, want: Action{Action: "fix-pr", Branch: branch, Failures: []Failure{{Workflow: "go", URL: url}}}},
		{name: "when Decide with waiting then returns waiting", state: Waiting{Branch: branch}, want: Action{Action: "wait", Branch: branch}},
		{name: "when Decide with stalled then returns waiting", state: Stalled{Branch: branch}, want: Action{Action: "wait", Branch: branch}},
		{name: "when Decide with passed then returns merging", state: Passed{Branch: branch}, want: Action{Action: "merge", Branch: branch}},
		{name: "when Decide with merged then returns merged", state: Merged{Branch: branch}, want: Action{Action: "merged", Branch: branch}},
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
