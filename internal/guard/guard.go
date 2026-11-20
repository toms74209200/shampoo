package guard

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
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
	settingsData, err := files.Optional(filepath.Join(repo.Root, ".shampoo", "settings.toml"))
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	options, err := settings.Parse(settingsData)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	branch, err := repo.Branch()
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	unmerged, err := repo.Unmerged()
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	timeBudget, timeStop := options.TimeBudget, options.TimeStop
	window := max(timeBudget.Runs, timeStop.Runs)
	for _, r := range rules {
		if checked, ok := r.(rule.Checked); ok {
			for _, transition := range checked.Transitions {
				window = max(window, checked.Budget[transition].Runs, checked.Stop[transition].Runs)
			}
		}
	}
	ledgerPath := filepath.Join(repo.CommonDir, "shampoo", "ledger.jsonl")
	var compaction budget.Compaction
	err = files.Update(ledgerPath, func(data []byte) ([]byte, error) {
		compaction = budget.Compact(budget.ParseLedger(data), append(unmerged, branch), window)
		var kept []byte
		for _, line := range compaction.Kept {
			encoded, err := json.Marshal(line)
			if err != nil {
				return nil, err
			}
			kept = append(append(kept, encoded...), '\n')
		}
		return kept, nil
	})
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	if len(compaction.Summaries) > 0 {
		err = files.Update(filepath.Join(repo.CommonDir, "shampoo", "stats.jsonl"), func(data []byte) ([]byte, error) {
			for _, summary := range compaction.Summaries {
				encoded, err := json.Marshal(summary)
				if err != nil {
					return nil, err
				}
				data = append(append(data, encoded...), '\n')
			}
			return data, nil
		})
		if err != nil {
			return rule.WorkflowFailure, "", err
		}
	}
	lines := compaction.Kept
	tree, err := repo.Snapshot()
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	attemptPath := filepath.Join(repo.GitDir, "shampoo", string(transition)+"-attempt.json")
	attemptData, err := files.Optional(attemptPath)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	var last attempt
	if err := json.Unmarshal(attemptData, &last); err == nil && last.Tree == tree {
		return last.Verdict, last.Report, nil
	}
	var base string
	if transition == rule.Task {
		base, err = repo.Base()
	} else {
		base, err = repo.Passed(transition)
	}
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	changed, err := repo.Changed(base, tree)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	selected, order := config.Plan(transition, changed)
	run := rand.Text()
	verdict := rule.Pass
	report := ""
checks:
	for _, check := range order {
		started := time.Now()
		code, output, err := shell.Run(repo.Root, check.Command)
		if err != nil {
			return rule.WorkflowFailure, report, fmt.Errorf("check %s: %w", check.ID, err)
		}
		millis := time.Since(started).Milliseconds()
		exceeded := false
		judgedByExit := false
		for _, checked := range selected {
			if checked.Check != check.ID {
				continue
			}
			var outcome rule.Outcome
			var detail string
			switch judge := checked.Judge.(type) {
			case rule.ByExit:
				outcome = judge.Exit.Outcome(code)
				detail = fmt.Sprintf("check %s: %s exited %d\n", check.ID, check.Command, code)
				judgedByExit = judgedByExit || outcome != rule.Pass
			case rule.ByCommand:
				judgeCode, judgeOutput, err := shell.Run(repo.Root, judge.Command)
				if err != nil {
					return rule.WorkflowFailure, report, fmt.Errorf("%s: judge: %w", checked.ID, err)
				}
				outcome = judge.Exit.Outcome(judgeCode)
				detail = fmt.Sprintf("judge: %s exited %d\n%s", judge.Command, judgeCode, judgeOutput)
			}
			line := budget.Line{Run: run, Task: branch, Transition: string(transition), Rule: checked.ID, Check: check.ID, Outcome: outcome, Millis: millis}
			encoded, err := json.Marshal(line)
			if err != nil {
				return rule.WorkflowFailure, report, err
			}
			err = files.Update(ledgerPath, func(data []byte) ([]byte, error) {
				return append(append(data, encoded...), '\n'), nil
			})
			if err != nil {
				return rule.WorkflowFailure, report, err
			}
			lines = append(lines, line)
			if outcome != rule.Pass {
				report += fmt.Sprintf("%s: %s\n%s, %s", checked.ID, checked.Statement, outcome, detail)
			}
			if verdict != rule.WorkflowFailure && (outcome == rule.TaskFailure || outcome == rule.WorkflowFailure) {
				verdict = outcome
			}
			if outcome != rule.TaskFailure && outcome != rule.Budget {
				continue
			}
			allowed, stop := checked.Budget[transition], checked.Stop[transition]
			if used := budget.Violations(lines, checked.ID, transition, allowed.Runs); allowed.Violations > 0 && used >= allowed.Violations {
				report += fmt.Sprintf("%s burns its budget at %s: %d violations in the latest %d runs, of %d allowed\n", checked.ID, transition, used, allowed.Runs, allowed.Violations)
			}
			if recent := budget.Violations(lines, checked.ID, transition, stop.Runs); allowed.Violations == 0 || recent >= stop.Violations {
				report += fmt.Sprintf("%s exceeds its budget at %s: %d violations in the latest %d runs\n", checked.ID, transition, recent, stop.Runs)
				exceeded = true
			}
		}
		limit := int64(options.MaxSeconds[string(check.Size)]) * 1000
		if used := budget.Overruns(lines, check.ID, limit, timeBudget.Runs); used >= timeBudget.Violations {
			report += fmt.Sprintf("check %s burns its time budget: %d runs over %ds of %s checks in the latest %d runs, of %d allowed\n", check.ID, used, limit/1000, check.Size, timeBudget.Runs, timeBudget.Violations)
		}
		if recent := budget.Overruns(lines, check.ID, limit, timeStop.Runs); recent >= timeStop.Violations {
			report += fmt.Sprintf("check %s exceeds its time budget: %d runs over %ds of %s checks in the latest %d runs: reorganize the check, such as by splitting it or moving it to a later transition\n", check.ID, recent, limit/1000, check.Size, timeStop.Runs)
			exceeded = true
		}
		if judgedByExit {
			report += fmt.Sprintf("output of check %s:\n%s", check.ID, output)
		}
		if exceeded {
			repair := budget.Line{Run: run, Task: branch, Transition: string(transition), Outcome: rule.Exceeded}
			encoded, err := json.Marshal(repair)
			if err != nil {
				return rule.WorkflowFailure, report, err
			}
			err = files.Update(ledgerPath, func(data []byte) ([]byte, error) {
				return append(append(data, encoded...), '\n'), nil
			})
			if err != nil {
				return rule.WorkflowFailure, report, err
			}
			lines = append(lines, repair)
			verdict = rule.Exceeded
			report += "repair the harness: find whether the work or the check causes the repeated violations, before going on with the task\n"
			if repairs := budget.Repairs(lines, branch); repairs > options.RepairAttempts {
				verdict = rule.Handed
				report += fmt.Sprintf("failed repairs in %s exceed their limit of %d: hand the task to a human\n", branch, options.RepairAttempts)
			}
			break checks
		}
	}
	if verdict == rule.Pass {
		if err := repo.Record(transition, tree); err != nil {
			return rule.WorkflowFailure, report, err
		}
	}
	encoded, err := json.Marshal(attempt{Tree: tree, Verdict: verdict, Report: report})
	if err != nil {
		return rule.WorkflowFailure, report, err
	}
	if err := files.Update(attemptPath, func([]byte) ([]byte, error) { return encoded, nil }); err != nil {
		return rule.WorkflowFailure, report, err
	}
	return verdict, report, nil
}

type attempt struct {
	Tree    string       `json:"tree"`
	Verdict rule.Outcome `json:"verdict"`
	Report  string       `json:"report"`
}

func Last(dir string, transition rule.Transition) (rule.Outcome, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return "", err
	}
	tree, err := repo.Snapshot()
	if err != nil {
		return "", err
	}
	data, err := files.Optional(filepath.Join(repo.GitDir, "shampoo", string(transition)+"-attempt.json"))
	if err != nil {
		return "", err
	}
	var last attempt
	if err := json.Unmarshal(data, &last); err != nil || last.Tree != tree {
		return "", nil
	}
	return last.Verdict, nil
}
