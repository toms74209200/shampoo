package guard

import (
	"fmt"
	"path/filepath"
	"slices"

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
	outcomes := map[string]rule.Outcome{}
	for _, check := range order {
		if slices.ContainsFunc(check.Needs, func(need string) bool { return outcomes[need] != rule.Pass }) {
			report += fmt.Sprintf("check %s: not run, since a check it needs did not pass\n", check.ID)
			continue
		}
		code, output, err := shell.Run(repo.Root, check.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("check %s: %w", check.ID, err)
		}
		outcome := check.Exit.Outcome(code)
		outcomes[check.ID] = outcome
		if outcome != rule.Pass {
			report += fmt.Sprintf("check %s: %s exited %d: %s\n%s", check.ID, check.Command, code, outcome, output)
			if verdict != rule.WorkflowFailure {
				verdict = outcome
			}
		}
	}
	for _, checked := range selected {
		if outcomes[checked.Check] != rule.Pass {
			report += fmt.Sprintf("%s: %s\ncheck %s did not pass\n", checked.ID, checked.Statement, checked.Check)
			continue
		}
		judge, ok := checked.Judge.(rule.Judged)
		if !ok {
			continue
		}
		code, output, err := shell.Run(repo.Root, judge.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("%s: judge: %w", checked.ID, err)
		}
		switch outcome := judge.Exit.Outcome(code); outcome {
		case rule.TaskFailure, rule.WorkflowFailure:
			report += fmt.Sprintf("%s: %s\n%s exited %d: %s\n%s", checked.ID, checked.Statement, judge.Command, code, outcome, output)
			if verdict != rule.WorkflowFailure {
				verdict = outcome
			}
		case rule.Budget:
			report += fmt.Sprintf("%s: %s\n%s exited %d: charges the budget of %s\n%s", checked.ID, checked.Statement, judge.Command, code, checked.ID, output)
		}
	}
	if verdict == rule.Pass {
		if err := repo.Record(transition, tree); err != nil {
			return rule.WorkflowFailure, report, err
		}
	}
	return verdict, report, nil
}
