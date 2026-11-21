package next

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/review"
	"github.com/toms74209200/shampoo/internal/rule"
)

type Observation struct {
	Source       string
	SourceFound  bool
	Branch       string
	Default      bool
	Questions    []string
	MaxOpen      int
	Escalations  []string
	Due          budget.Due
	Verdict      rule.Outcome
	Tree         string
	Policies     []policy.Policy
	Requirements map[string]any
	Record       *Record
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
type State interface{ state() }

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

type Editing struct{ Branch string }

type Repairing struct{ Branch string }

type Handed struct{ Branch string }

type Done struct {
	Branch       string
	Policies     []string
	Requirements map[string]any
}

type Ready struct {
	Branch   string
	Findings []review.Finding
}

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

func (NoSource) state()          {}
func (Idle) state()              {}
func (Blocked) state()           {}
func (ReviewingWorkflow) state() {}
func (Escalating) state()        {}
func (Editing) state()           {}
func (Repairing) state()         {}
func (Handed) state()            {}
func (Done) state()              {}
func (Ready) state()             {}
func (Failing) state()           {}
func (Stalled) state()           {}
func (Waiting) state()           {}
func (Passed) state()            {}
func (Merged) state()            {}

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

func Parse(o Observation) State {
	if o.Source == "" || (!strings.Contains(o.Source, "://") && !o.SourceFound) {
		return NoSource{Source: o.Source}
	}
	if o.Default && len(o.Escalations) > 0 {
		return Blocked{Questions: o.Escalations}
	}
	if o.Default && len(o.Questions) > o.MaxOpen {
		return Blocked{Questions: o.Questions}
	}
	if o.Default {
		switch due := o.Due.(type) {
		case budget.Review:
			return ReviewingWorkflow{Asked: due.Asked, Repaired: due.Repaired}
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
		return Idle{Source: o.Source}
	}
	pr := o.PullRequest
	if pr != nil && pr.State == "MERGED" {
		return Merged{Branch: o.Branch}
	}
	if pr != nil && pr.State == "CLOSED" {
		return Handed{Branch: o.Branch}
	}
	switch o.Verdict {
	case rule.Exceeded:
		return Repairing{Branch: o.Branch}
	case rule.Handed:
		return Handed{Branch: o.Branch}
	}
	if o.Verdict != rule.Pass {
		return Editing{Branch: o.Branch}
	}
	if pr == nil || !pr.Pushed {
		var marked []string
		for _, p := range o.Policies {
			if p.Review {
				marked = append(marked, p.ID)
			}
		}
		mine := o.Record != nil && o.Record.Branch == o.Branch
		switch {
		case !mine && len(marked) > 0:
			return Done{Branch: o.Branch, Policies: marked, Requirements: o.Requirements}
		case mine && o.Record.Tree == o.Tree:
			return Ready{Branch: o.Branch, Findings: o.Record.Findings}
		}
		return Ready{Branch: o.Branch}
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
		return Failing{Branch: o.Branch, Failures: failures}
	case waiting && o.Now.Sub(pr.Latest) >= o.MaxWait:
		return Stalled{Branch: o.Branch, Since: pr.Latest, MaxWait: o.MaxWait}
	case waiting:
		return Waiting{Branch: o.Branch}
	}
	return Passed{Branch: o.Branch}
}

func Decide(s State) Action {
	switch s := s.(type) {
	case NoSource:
		return Action{Action: "set-task-source", Source: s.Source}
	case Idle:
		return Action{Action: "take-task", Source: s.Source}
	case Blocked:
		return Action{Action: "wait-for-human", Questions: s.Questions}
	case ReviewingWorkflow:
		return Action{Action: "review-workflow", Asked: s.Asked, Repaired: s.Repaired}
	case Escalating:
		return Action{Action: "wait-for-human"}
	case Editing:
		return Action{Action: "continue", Branch: s.Branch}
	case Repairing:
		return Action{Action: "repair-harness", Branch: s.Branch}
	case Handed:
		return Action{Action: "hand-to-human", Branch: s.Branch}
	case Done:
		return Action{Action: "review", Branch: s.Branch, Policies: s.Policies, Requirements: s.Requirements}
	case Ready:
		return Action{Action: "open-pr", Branch: s.Branch, Findings: s.Findings}
	case Failing:
		return Action{Action: "fix-pr", Branch: s.Branch, Failures: s.Failures}
	case Waiting:
		return Action{Action: "wait", Branch: s.Branch}
	case Stalled:
		return Action{Action: "wait", Branch: s.Branch}
	case Passed:
		return Action{Action: "merge", Branch: s.Branch}
	case Merged:
		return Action{Action: "merged", Branch: s.Branch}
	}
	return Action{}
}
