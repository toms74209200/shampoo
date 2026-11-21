package judgment

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
)

type Result struct {
	Code   int
	Output []byte
	Millis int64
}

type World struct {
	Run    func(command string) (Result, error)
	Remove func(path string) error
	Read   func(path string) ([]byte, error)
	Append func(line budget.Line) error
	Pass   func() error
}

type Session struct {
	Run        string
	Task       string
	Transition rule.Transition
	Selected   []rule.Checked
	Order      []rule.Check
	Lines      []budget.Line
	Options    settings.Settings
}

func Judge(s Session, w World) (rule.Outcome, string, error) {
	lines := s.Lines
	var outcomes []rule.Outcome
	report := ""
	failures := budget.Limits{Budget: s.Options.FailureBudget, Stop: s.Options.FailureStop}
	time := budget.Limits{Budget: s.Options.TimeBudget, Stop: s.Options.TimeStop}
	for _, check := range s.Order {
		var judged []rule.Checked
		for _, checked := range s.Selected {
			if checked.Check == check.ID {
				judged = append(judged, checked)
			}
		}
		for _, checked := range judged {
			if checked.Exceptions == nil {
				continue
			}
			if err := w.Remove(checked.Exceptions.File); err != nil {
				return rule.WorkflowFailure, report, err
			}
		}
		result, err := w.Run(check.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("check %s: %w", check.ID, err)
		}
		exceeded := false
		judgedByExit := false
		for _, checked := range judged {
			var outcome rule.Outcome
			var detail string
			switch judge := checked.Judge.(type) {
			case rule.ByExit:
				outcome = judge.Exit.Outcome(result.Code)
				detail = fmt.Sprintf("check %s: %s exited %d\n", check.ID, check.Command, result.Code)
				judgedByExit = judgedByExit || outcome != rule.Pass
			case rule.ByCommand:
				judgment, err := w.Run(judge.Command)
				if err != nil {
					return rule.WorkflowFailure, report, fmt.Errorf("%s: judge: %w", checked.ID, err)
				}
				outcome = judge.Exit.Outcome(judgment.Code)
				detail = fmt.Sprintf("judge: %s exited %d\n%s", judge.Command, judgment.Code, judgment.Output)
			}
			var exceptions *int
			if checked.Exceptions != nil {
				data, err := w.Read(checked.Exceptions.File)
				count := len(strings.FieldsFunc(string(data), func(r rune) bool { return r == '\n' }))
				exceptions = &count
				if err != nil {
					outcome = rule.WorkflowFailure
					detail = fmt.Sprintf("check %s did not write its exceptions to %s: %v\n", check.ID, checked.Exceptions.File, err)
					exceptions = nil
				}
			}
			line := budget.Line{Run: s.Run, Task: s.Task, Transition: string(s.Transition), Rule: checked.ID, Check: check.ID, Outcome: outcome, Millis: result.Millis, Exceptions: exceptions}
			if err := w.Append(line); err != nil {
				return rule.WorkflowFailure, report, err
			}
			lines = append(lines, line)
			if outcome != rule.Pass {
				report += fmt.Sprintf("%s: %s\n%s, %s", checked.ID, checked.Statement, outcome, detail)
			}
			outcomes = append(outcomes, outcome)
			charge := budget.ChargeRule(lines, checked, s.Transition, outcome, exceptions, failures)
			report += charge.Report
			exceeded = exceeded || charge.Exceeded
		}
		charge := budget.ChargeTime(lines, check, s.Options.MaxSeconds[string(check.Size)], time)
		report += charge.Report
		exceeded = exceeded || charge.Exceeded
		if judgedByExit {
			report += fmt.Sprintf("output of check %s:\n%s", check.ID, result.Output)
		}
		if !exceeded {
			continue
		}
		repair := budget.Line{Run: s.Run, Task: s.Task, Transition: string(s.Transition), Outcome: rule.Exceeded}
		if err := w.Append(repair); err != nil {
			return rule.WorkflowFailure, report, err
		}
		lines = append(lines, repair)
		report += "repair the harness: find whether the work or the check causes the repeated violations, before going on with the task\n"
		if repairs := budget.Repairs(lines, s.Task); repairs > s.Options.RepairAttempts {
			report += fmt.Sprintf("failed repairs in %s exceed their limit of %d: hand the task to a human\n", s.Task, s.Options.RepairAttempts)
			return rule.Handed, report, nil
		}
		return rule.Exceeded, report, nil
	}
	verdict := rule.Pass
	switch {
	case slices.Contains(outcomes, rule.WorkflowFailure):
		verdict = rule.WorkflowFailure
	case slices.Contains(outcomes, rule.TaskFailure):
		verdict = rule.TaskFailure
	}
	if verdict == rule.Pass {
		if err := w.Pass(); err != nil {
			return rule.WorkflowFailure, report, err
		}
	}
	return verdict, report, nil
}

type Attempt struct {
	Tree    string       `json:"tree"`
	Verdict rule.Outcome `json:"verdict"`
	Report  string       `json:"report"`
}

//sumtype:decl
type Earlier interface{ earlier() }

type Cached struct {
	Verdict rule.Outcome
	Report  string
}

type Fresh struct{}

func (Cached) earlier() {}
func (Fresh) earlier()  {}

func ParseAttempt(data []byte, tree string) Earlier {
	var last Attempt
	if err := json.Unmarshal(data, &last); err != nil || last.Tree != tree {
		return Fresh{}
	}
	return Cached{Verdict: last.Verdict, Report: last.Report}
}
