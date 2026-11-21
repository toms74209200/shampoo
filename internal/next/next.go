package next

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/question"
	"github.com/toms74209200/shampoo/internal/review"
	"github.com/toms74209200/shampoo/internal/rule"
)

type Observation struct {
	Source       string
	SourceFound  bool
	Branch       string
	Default      bool
	MaxOpen      int
	Due          budget.Due
	Verdict      rule.Outcome
	Tree         string
	Policies     []policy.Policy
	Requirements map[string]any
	Record       *Record
	Convergence  budget.Convergence
	PullRequest  *PullRequest
	Now          time.Time
	MaxWait      time.Duration
}

type Record struct {
	Branch   string           `json:"branch"`
	Tree     string           `json:"tree"`
	Findings []review.Finding `json:"findings"`
}

type PullRequest struct {
	State    string
	Pushed   bool
	Commits  []string
	Latest   time.Time
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

//sumtype:decl
type Observed interface{ observed() }

//sumtype:decl
type State interface{ state() }

type Settled struct{ State State }

type Starting struct{}

type DueReview struct {
	Asked    []string
	Repaired []string
}

type NoSource struct{ Source string }

type Idle struct{ Source string }

type Blocked struct{ Questions []string }

type ReviewingWorkflow struct {
	Asked    []string
	Repaired []string
}

type Escalating struct {
	Title string
	Body  string
}

type Editing struct {
	Branch   string
	Findings []review.Finding
}

type Repairing struct {
	Branch   string
	Policies []string
}

type Handed struct{ Branch string }

type Done struct {
	Branch       string
	Policies     []string
	Requirements map[string]any
}

type Ready struct{ Branch string }

type Failing struct {
	Branch   string
	Failures []Failure
}

type Waiting struct{ Branch string }

type Stalled struct {
	Branch  string
	Since   time.Time
	MaxWait time.Duration
}

type Passed struct{ Branch string }

type Merged struct{ Branch string }

func (Idle) state()              {}
func (Blocked) state()           {}
func (ReviewingWorkflow) state() {}
func (Editing) state()           {}
func (Repairing) state()         {}
func (Handed) state()            {}
func (Done) state()              {}
func (Ready) state()             {}
func (Failing) state()           {}
func (Waiting) state()           {}
func (Merged) state()            {}

func (Settled) observed()    {}
func (Starting) observed()   {}
func (NoSource) observed()   {}
func (DueReview) observed()  {}
func (Escalating) observed() {}
func (Stalled) observed()    {}
func (Passed) observed()     {}

type Action struct {
	Action       string           `json:"action"`
	Source       string           `json:"source,omitempty"`
	Branch       string           `json:"branch,omitempty"`
	Failures     []Failure        `json:"failures,omitempty"`
	Policies     []string         `json:"policies,omitempty"`
	Requirements map[string]any   `json:"requirements,omitempty"`
	Findings     []review.Finding `json:"findings,omitempty"`
	Questions    []string         `json:"questions,omitempty"`
	Asked        []string         `json:"asked,omitempty"`
	Repaired     []string         `json:"repaired,omitempty"`
}

type Failure struct {
	Workflow string `json:"workflow"`
	URL      string `json:"url"`
}

func ParseRecord(data []byte) (*Record, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("review record: %w", err)
	}
	return &record, nil
}

func Parse(o Observation) Observed {
	if o.Source == "" || (!strings.Contains(o.Source, "://") && !o.SourceFound) {
		return NoSource{Source: o.Source}
	}
	if o.Default {
		return Starting{}
	}
	pr := o.PullRequest
	if pr != nil && pr.State == "MERGED" {
		return Settled{State: Merged{Branch: o.Branch}}
	}
	if pr != nil && pr.State == "CLOSED" {
		return Settled{State: Handed{Branch: o.Branch}}
	}
	switch o.Verdict {
	case rule.Exceeded:
		return Settled{State: Repairing{Branch: o.Branch}}
	case rule.Handed:
		return Settled{State: Handed{Branch: o.Branch}}
	}
	if o.Verdict != rule.Pass {
		return Settled{State: Editing{Branch: o.Branch}}
	}
	if pr == nil || !pr.Pushed {
		var marked []string
		for _, p := range o.Policies {
			if p.Review {
				marked = append(marked, p.ID)
			}
		}
		mine := o.Record != nil && o.Record.Branch == o.Branch
		found := mine && len(o.Record.Findings) > 0
		switch {
		case !mine && len(marked) > 0, found && o.Record.Tree != o.Tree:
			return Settled{State: Done{Branch: o.Branch, Policies: marked, Requirements: o.Requirements}}
		case found:
			switch o.Convergence.(type) {
			case budget.Exhausted:
				return Settled{State: Handed{Branch: o.Branch}}
			case budget.Diverging:
				var policies []string
				for _, finding := range o.Record.Findings {
					if !slices.Contains(policies, finding.Policy) {
						policies = append(policies, finding.Policy)
					}
				}
				return Settled{State: Repairing{Branch: o.Branch, Policies: policies}}
			case budget.Converging:
			}
			return Settled{State: Editing{Branch: o.Branch, Findings: o.Record.Findings}}
		}
		return Settled{State: Ready{Branch: o.Branch}}
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
		return Settled{State: Failing{Branch: o.Branch, Failures: failures}}
	case waiting && o.Now.Sub(pr.Latest) >= o.MaxWait:
		return Stalled{Branch: o.Branch, Since: pr.Latest, MaxWait: o.MaxWait}
	case waiting:
		return Settled{State: Waiting{Branch: o.Branch}}
	}
	return Passed{Branch: o.Branch}
}

func Start(o Observation, opened question.Open) Observed {
	if len(opened.Workflows) > 0 {
		return Settled{State: Blocked{Questions: opened.Workflows}}
	}
	if len(opened.Questions) > o.MaxOpen {
		return Settled{State: Blocked{Questions: opened.Questions}}
	}
	switch due := o.Due.(type) {
	case budget.Review:
		return DueReview{Asked: due.Asked, Repaired: due.Repaired}
	case budget.Escalate:
		body := "Workflow reviews keep failing: the tasks below still ask a human or repair the harness too often. Review the workflow, then close this issue.\n\nAsked a human:\n"
		for _, task := range due.Asked {
			body += "- " + task + "\n"
		}
		body += "\nRepaired the harness:\n"
		for _, task := range due.Repaired {
			body += "- " + task + "\n"
		}
		return Escalating{Title: "Review the workflow: reviews keep failing", Body: body}
	case budget.Steady:
	}
	return Settled{State: Idle{Source: o.Source}}
}

func Decide(s State) Action {
	switch s := s.(type) {
	case Idle:
		return Action{Action: "take-task", Source: s.Source}
	case Blocked:
		return Action{Action: "wait-for-human", Questions: s.Questions}
	case ReviewingWorkflow:
		return Action{Action: "review-workflow", Asked: s.Asked, Repaired: s.Repaired}
	case Editing:
		return Action{Action: "continue", Branch: s.Branch, Findings: s.Findings}
	case Repairing:
		return Action{Action: "repair-harness", Branch: s.Branch, Policies: s.Policies}
	case Handed:
		return Action{Action: "hand-to-human", Branch: s.Branch}
	case Done:
		return Action{Action: "review", Branch: s.Branch, Policies: s.Policies, Requirements: s.Requirements}
	case Ready:
		return Action{Action: "open-pr", Branch: s.Branch}
	case Failing:
		return Action{Action: "fix-pr", Branch: s.Branch, Failures: s.Failures}
	case Waiting:
		return Action{Action: "wait", Branch: s.Branch}
	case Merged:
		return Action{Action: "merged", Branch: s.Branch}
	}
	return Action{}
}
