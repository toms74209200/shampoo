package guard

import (
	"fmt"
	"path/filepath"

	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/shell"
)

func Run(dir string, transition rule.Transition) (rule.Outcome, string, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	ruleSources, err := files.Read(filepath.Join(repo.Root, ".shampoo", "rules"), ".toml")
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	checkSources, err := files.Read(filepath.Join(repo.Root, ".shampoo", "checks"), ".toml")
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	var rules []rule.Rule
	for _, source := range ruleSources {
		r, err := rule.Parse(source.Name, source.Data)
		if err != nil {
			return rule.WorkflowFailure, "", err
		}
		rules = append(rules, r)
	}
	var checks []rule.Check
	for _, source := range checkSources {
		check, err := rule.ParseCheck(source.Name, source.Data)
		if err != nil {
			return rule.WorkflowFailure, "", err
		}
		checks = append(checks, check)
	}
	config, err := rule.Link(rules, checks)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	tree, err := repo.Snapshot()
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	passed, err := repo.Passed(transition)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	changed, err := repo.Changed(passed, tree)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	selected, order := config.Plan(transition, changed)
	verdict := rule.Pass
	report := ""
	for _, check := range order {
		code, output, err := shell.Run(repo.Root, check.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("check %s: %w", check.ID, err)
		}
		exceeded := false
		for _, checked := range selected {
			if checked.Check != check.ID {
				continue
			}
			var outcome rule.Outcome
			var detail string
			switch judge := checked.Judge.(type) {
			case rule.ByExit:
				outcome = judge.Exit.Outcome(code)
				detail = fmt.Sprintf("check %s: %s exited %d\n%s", check.ID, check.Command, code, output)
			case rule.ByCommand:
				judgeCode, judgeOutput, err := shell.Run(repo.Root, judge.Command)
				if err != nil {
					return rule.WorkflowFailure, report, fmt.Errorf("%s: judge: %w", checked.ID, err)
				}
				outcome = judge.Exit.Outcome(judgeCode)
				detail = fmt.Sprintf("judge: %s exited %d\n%s", judge.Command, judgeCode, judgeOutput)
			}
			if outcome != rule.Pass {
				report += fmt.Sprintf("%s: %s\n%s, %s", checked.ID, checked.Statement, outcome, detail)
			}
			if verdict != rule.WorkflowFailure && (outcome == rule.TaskFailure || outcome == rule.WorkflowFailure) {
				verdict = outcome
			}
			if (outcome == rule.TaskFailure || outcome == rule.Budget) && checked.Budget == 0 {
				report += fmt.Sprintf("%s: exceeds its budget of 0\n", checked.ID)
				exceeded = true
			}
		}
		if exceeded {
			return rule.Exceeded, report, nil
		}
	}
	if verdict == rule.Pass {
		if err := repo.Record(transition, tree); err != nil {
			return rule.WorkflowFailure, report, err
		}
	}
	return verdict, report, nil
}
