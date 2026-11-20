package next

import (
	"slices"
	"time"

	"github.com/toms74209200/shampoo/internal/rule"
)

type State struct {
	Source      string
	Branch      string
	Default     bool
	Verdict     rule.Outcome
	PullRequest *PullRequest
}

type PullRequest struct {
	State    string
	Pushed   bool
	Commits  []string
	Expected []string
	Runs     []Run
}

type Run struct {
	Workflow   string
	Commit     string
	Status     string
	Conclusion string
	URL        string
	Created    time.Time
}

type Action struct {
	Action   string    `json:"action"`
	Source   string    `json:"source,omitempty"`
	Branch   string    `json:"branch,omitempty"`
	Failures []Failure `json:"failures,omitempty"`
}

type Failure struct {
	Workflow string `json:"workflow"`
	URL      string `json:"url"`
}

func Decide(s State) Action {
	if s.Default {
		return Action{Action: "take-task", Source: s.Source}
	}
	pr := s.PullRequest
	if pr != nil && pr.State == "MERGED" {
		return Action{Action: "merged", Branch: s.Branch}
	}
	if pr != nil && pr.State == "CLOSED" {
		return Action{Action: "hand-to-human", Branch: s.Branch}
	}
	switch s.Verdict {
	case rule.Pass:
		if pr == nil || !pr.Pushed {
			return Action{Action: "open-pr", Branch: s.Branch}
		}
	case rule.Exceeded:
		return Action{Action: "repair-harness", Branch: s.Branch}
	case rule.Handed:
		return Action{Action: "hand-to-human", Branch: s.Branch}
	default:
		return Action{Action: "continue", Branch: s.Branch}
	}
	var failures []Failure
	waiting := false
	for _, workflow := range pr.Expected {
		var latest *Run
		for i, run := range pr.Runs {
			if run.Workflow == workflow && slices.Contains(pr.Commits, run.Commit) && run.Conclusion != "action_required" && (latest == nil || run.Created.After(latest.Created)) {
				latest = &pr.Runs[i]
			}
		}
		switch {
		case latest == nil || latest.Status != "completed":
			waiting = true
		case !slices.Contains([]string{"success", "neutral", "skipped"}, latest.Conclusion):
			failures = append(failures, Failure{Workflow: workflow, URL: latest.URL})
		}
	}
	switch {
	case len(failures) > 0:
		return Action{Action: "fix-pr", Branch: s.Branch, Failures: failures}
	case waiting:
		return Action{Action: "wait", Branch: s.Branch}
	}
	return Action{Action: "merge", Branch: s.Branch}
}
