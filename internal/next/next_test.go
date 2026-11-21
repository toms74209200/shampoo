//go:build small

package next

import (
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/policy"
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
		{name: "when Parse with no task source then returns no source", observation: Observation{Source: "", SourceFound: false, Default: true}, want: NoSource{Source: ""}},
		{name: "when Parse with a task file not found then returns no source", observation: Observation{Source: "todo.md", SourceFound: false, Default: true}, want: NoSource{Source: "todo.md"}},
		{name: "when Parse with a task URL then returns idle without looking for a file", observation: Observation{Source: "https://github.example/issues", SourceFound: false, Default: true}, want: Idle{Source: "https://github.example/issues"}},
		{name: "when Parse with a task branch and a review policy not marked for review then returns ready", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "layout"}}}, want: Ready{Branch: branch}},
		{name: "when Parse with the default branch then returns idle with the source", observation: Observation{Source: source, SourceFound: true, Branch: branch, Default: true, Verdict: rule.Outcome("pass")}, want: Idle{Source: source}},
		{name: "when Parse with the default branch and open questions within the limit then returns idle", observation: Observation{Source: source, SourceFound: true, Default: true, Questions: []string{url}, MaxOpen: 1}, want: Idle{Source: source}},
		{name: "when Parse with the default branch and more open questions than the limit then returns blocked by them", observation: Observation{Source: source, SourceFound: true, Default: true, Questions: []string{url, other}, MaxOpen: 1}, want: Blocked{Questions: []string{url, other}}},
		{name: "when Parse with the default branch and a steady workflow then returns idle", observation: Observation{Source: source, SourceFound: true, Default: true, Due: budget.Steady{}}, want: Idle{Source: source}},
		{name: "when Parse with the default branch and a workflow due for review then returns reviewing it with the tasks", observation: Observation{Source: source, SourceFound: true, Default: true, Due: budget.Review{Asked: []string{first}, Repaired: []string{second}}}, want: ReviewingWorkflow{Asked: []string{first}, Repaired: []string{second}}},
		{name: "when Parse with the default branch and failing workflow reviews then returns escalating with the tasks", observation: Observation{Source: source, SourceFound: true, Default: true, Due: budget.Escalate{Asked: []string{"a"}, Repaired: []string{"b"}}}, want: Escalating{Title: "Review the workflow: reviews keep failing", Body: "Workflow reviews keep failing: the tasks below still ask a human or repair the harness too often. Review the workflow, then close this issue.\n\nAsked a human:\n- a\n\nRepaired the harness:\n- b\n"}},
		{name: "when Parse with the default branch and an open escalation then returns blocked by it", observation: Observation{Source: source, SourceFound: true, Default: true, Escalations: []string{url}, Due: budget.Review{Asked: []string{first}}}, want: Blocked{Questions: []string{url}}},
		{name: "when Parse with a task branch and a workflow due for review then returns editing", observation: Observation{Source: source, SourceFound: true, Branch: branch, Escalations: []string{url}, Due: budget.Review{Asked: []string{first}}}, want: Editing{Branch: branch}},
		{name: "when Parse with a task branch and more open questions than the limit then returns editing", observation: Observation{Source: source, SourceFound: true, Branch: branch, Questions: []string{url, other}, MaxOpen: 1}, want: Editing{Branch: branch}},
		{name: "when Parse with a passed task guard then returns ready", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass")}, want: Ready{Branch: branch}},
		{name: "when Parse with a budget exceeded then returns repairing", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("budget-exceeded")}, want: Repairing{Branch: branch}},
		{name: "when Parse with repairs beyond their limit then returns handed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("repair-exceeded")}, want: Handed{Branch: branch}},
		{name: "when Parse with a task failure then returns editing", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("task-failure")}, want: Editing{Branch: branch}},
		{name: "when Parse with no judgment of the current tree then returns editing", observation: Observation{Source: source, SourceFound: true, Branch: branch}, want: Editing{Branch: branch}},
		{name: "when Parse with a passed task guard and policies for review then returns done with them and the requirements", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "layout"}, {ID: "functions", Review: true}}, Requirements: map[string]any{"model": source}}, want: Done{Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}},
		{name: "when Parse with a review with findings on the current tree then returns editing with them", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: branch, Tree: tree, Findings: []review.Finding{{Policy: "functions", Comment: url}}}, Convergence: budget.Converging{}}, want: Editing{Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}},
		{name: "when Parse with a review with findings on an earlier tree then returns done to review again", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: branch, Tree: other, Findings: []review.Finding{{Policy: "functions", Comment: url}}}}, want: Done{Branch: branch, Policies: []string{"functions"}}},
		{name: "when Parse with a review without findings on an earlier tree then returns ready", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: branch, Tree: other}}, want: Ready{Branch: branch}},
		{name: "when Parse with findings that diverge then returns repairing their policies once each", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: branch, Tree: tree, Findings: []review.Finding{{Policy: "functions", Comment: url}, {Policy: "tests", Comment: url}, {Policy: "functions", Comment: older}}}, Convergence: budget.Diverging{}}, want: Repairing{Branch: branch, Policies: []string{"functions", "tests"}}},
		{name: "when Parse with findings that diverge beyond the repair attempts then returns handed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: branch, Tree: tree, Findings: []review.Finding{{Policy: "functions", Comment: url}}}, Convergence: budget.Exhausted{}}, want: Handed{Branch: branch}},
		{name: "when Parse with a review of another task then returns done", observation: Observation{Source: source, SourceFound: true, Branch: branch, Tree: tree, Verdict: rule.Outcome("pass"), Policies: []policy.Policy{{ID: "functions", Review: true}}, Record: &Record{Branch: other, Tree: tree}}, want: Done{Branch: branch, Policies: []string{"functions"}}},
		{name: "when Parse with a passed guard on commits the pull request lacks then returns ready", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", false, nil)}, want: Ready{Branch: branch}},
		{name: "when Parse with a task failure on an open pull request then returns editing", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("task-failure"), PullRequest: pr("OPEN", true, nil)}, want: Editing{Branch: branch}},
		{name: "when Parse with a merged pull request then returns merged", observation: Observation{Source: source, SourceFound: true, Branch: branch, PullRequest: pr("MERGED", true, nil)}, want: Merged{Branch: branch}},
		{name: "when Parse with a closed pull request then returns handed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("CLOSED", true, nil)}, want: Handed{Branch: branch}},
		{name: "when Parse with no run of an expected workflow then returns waiting", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}), Now: start.Add(59 * time.Minute), MaxWait: time.Hour}, want: Waiting{Branch: branch}},
		{name: "when Parse with no run of an expected workflow for the longest wait then returns stalled since the latest commit", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}), Now: start.Add(time.Hour), MaxWait: time.Hour}, want: Stalled{Branch: branch, Since: start, MaxWait: time.Hour}},
		{name: "when Parse with an expected workflow running then returns waiting", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "in_progress", URL: url, Created: start}), Now: start, MaxWait: time.Hour}, want: Waiting{Branch: branch}},
		{name: "when Parse with an expected workflow failed then returns failing with its run", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "commit"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "failure", URL: url, Created: start})}, want: Failing{Branch: branch, Failures: []Failure{{Workflow: "go", URL: url}}}},
		{name: "when Parse with every expected workflow passed then returns passed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go", "docs", "lint"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "docs", Commit: second, Status: "completed", Conclusion: "skipped", Created: start}, Run{Workflow: "lint", Commit: second, Status: "completed", Conclusion: "neutral", Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a run on an earlier commit then returns passed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a later run that did not execute then returns passed by the earlier run", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "success", Created: start}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "action_required", Created: start.Add(time.Minute)})}, want: Passed{Branch: branch}},
		{name: "when Parse with a failed run before a passed one then returns passed", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: second, Status: "completed", Conclusion: "success", URL: other, Created: start.Add(time.Minute)}, Run{Workflow: "go", Commit: first, Status: "completed", Conclusion: "failure", URL: older, Created: start})}, want: Passed{Branch: branch}},
		{name: "when Parse with a run on a commit outside the pull request then returns waiting", observation: Observation{Source: source, SourceFound: true, Branch: branch, Verdict: rule.Outcome("pass"), PullRequest: pr("OPEN", true, []string{"go"}, Run{Workflow: "go", Commit: elsewhere, Status: "completed", Conclusion: "success", Created: start}), Now: start, MaxWait: time.Hour}, want: Waiting{Branch: branch}},
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
		{name: "when Decide with no source then returns setting the task source", state: NoSource{Source: source}, want: Action{Action: "set-task-source", Source: source}},
		{name: "when Decide with reviewing the workflow then returns reviewing it with the tasks", state: ReviewingWorkflow{Asked: []string{url}, Repaired: []string{branch}}, want: Action{Action: "review-workflow", Asked: []string{url}, Repaired: []string{branch}}},
		{name: "when Decide with escalating then returns waiting for a human", state: Escalating{Title: url, Body: branch}, want: Action{Action: "wait-for-human"}},
		{name: "when Decide with idle then returns taking a task from the source", state: Idle{Source: source}, want: Action{Action: "take-task", Source: source}},
		{name: "when Decide with blocked then returns waiting for a human with the questions", state: Blocked{Questions: []string{url}}, want: Action{Action: "wait-for-human", Questions: []string{url}}},
		{name: "when Decide with editing then returns going on with the task and its findings", state: Editing{Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}, want: Action{Action: "continue", Branch: branch, Findings: []review.Finding{{Policy: "functions", Comment: url}}}},
		{name: "when Decide with repairing then returns repairing the harness with the policies", state: Repairing{Branch: branch, Policies: []string{"functions"}}, want: Action{Action: "repair-harness", Branch: branch, Policies: []string{"functions"}}},
		{name: "when Decide with handed then returns handing the task to a human", state: Handed{Branch: branch}, want: Action{Action: "hand-to-human", Branch: branch}},
		{name: "when Decide with done then returns reviewing with the policies and the requirements", state: Done{Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}, want: Action{Action: "review", Branch: branch, Policies: []string{"functions"}, Requirements: map[string]any{"model": source}}},
		{name: "when Decide with ready then returns opening a pull request", state: Ready{Branch: branch}, want: Action{Action: "open-pr", Branch: branch}},
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

func TestParseRecord(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *Record
		wantErr bool
	}{
		{name: "when ParseRecord with no data then returns no record", data: "", want: nil},
		{name: "when ParseRecord with a record then returns it", data: `{"branch": "task-a", "tree": "t1", "findings": [{"policy": "functions", "comment": "a wrapper"}]}`, want: &Record{Branch: "task-a", Tree: "t1", Findings: []review.Finding{{Policy: "functions", Comment: "a wrapper"}}}},
		{name: "when ParseRecord with a record of no branch then returns it with none", data: `{"tree": "t1", "findings": []}`, want: &Record{Tree: "t1", Findings: []review.Finding{}}},
		{name: "when ParseRecord with invalid JSON then returns an error", data: `{"tree"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRecord([]byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
