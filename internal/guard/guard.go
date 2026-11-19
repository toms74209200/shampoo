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
	sources, err := files.Read(filepath.Join(repo.Root, ".shampoo", "rules"), ".toml")
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	var rules []rule.Rule
	for _, source := range sources {
		r, err := rule.Parse(source.Name, source.Data)
		if err != nil {
			return rule.WorkflowFailure, "", err
		}
		rules = append(rules, r)
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
	verdict := rule.Pass
	report := ""
	for _, checked := range rule.Select(rules, transition, changed) {
		code, output, err := shell.Run(repo.Root, checked.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("%s: %w", checked.ID, err)
		}
		switch outcome := checked.Exit.Outcome(code); outcome {
		case rule.TaskFailure, rule.WorkflowFailure:
			report += fmt.Sprintf("%s: %s\n%s exited %d: %s\n%s", checked.ID, checked.Statement, checked.Command, code, outcome, output)
			if verdict != rule.WorkflowFailure {
				verdict = outcome
			}
		case rule.Budget:
			report += fmt.Sprintf("%s: %s\n%s exited %d: charges the budget of %s\n%s", checked.ID, checked.Statement, checked.Command, code, checked.ID, output)
		}
	}
	if verdict == rule.Pass {
		if err := repo.Record(transition, tree); err != nil {
			return rule.WorkflowFailure, report, err
		}
	}
	return verdict, report, nil
}
