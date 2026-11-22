package guard

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/hook"
	"github.com/toms74209200/shampoo/internal/judgment"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
	"github.com/toms74209200/shampoo/internal/shell"
	"github.com/toms74209200/shampoo/internal/usage"
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
	window := max(config.Window(), options.TimeBudget.Runs, options.TimeStop.Runs, options.FailureBudget.Runs, options.FailureStop.Runs)
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
	switch earlier := judgment.ParseAttempt(attemptData, tree).(type) {
	case judgment.Cached:
		return earlier.Verdict, earlier.Report, nil
	case judgment.Fresh:
	}
	passed, err := repo.Passed(rule.Edit)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	mergeBase, err := repo.Base()
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	base := map[rule.Transition]string{rule.Edit: passed, rule.Task: mergeBase}[transition]
	changed, err := repo.Changed(base, tree)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	selected, order := config.Plan(transition, changed)
	ledger := func(line budget.Line) error {
		encoded, err := json.Marshal(line)
		if err != nil {
			return err
		}
		return files.Update(ledgerPath, func(data []byte) ([]byte, error) {
			return append(append(data, encoded...), '\n'), nil
		})
	}
	world := judgment.World{
		Run: func(command string) (judgment.Result, error) {
			started := time.Now()
			code, output, err := shell.Run(repo.Root, command)
			return judgment.Result{Code: code, Output: output, Millis: time.Since(started).Milliseconds()}, err
		},
		Remove: func(path string) error {
			if err := os.Remove(filepath.Join(repo.Root, path)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			return nil
		},
		Read:   func(path string) ([]byte, error) { return os.ReadFile(filepath.Join(repo.Root, path)) },
		Append: ledger,
		Pass:   func() error { return repo.Record(transition, tree) },
	}
	session := judgment.Session{Run: rand.Text(), Task: branch, Transition: transition, Selected: selected, Order: order, Lines: lines, Options: options}
	verdict, report, err := judgment.Judge(session, world)
	if err != nil {
		return rule.WorkflowFailure, report, err
	}
	encoded, err := json.Marshal(judgment.Attempt{Tree: tree, Verdict: verdict, Report: report})
	if err != nil {
		return rule.WorkflowFailure, report, err
	}
	if err := files.Update(attemptPath, func([]byte) ([]byte, error) { return encoded, nil }); err != nil {
		return rule.WorkflowFailure, report, err
	}
	return verdict, report, nil
}

func Usage(dir string, named hook.Transcript) error {
	var path string
	switch t := named.(type) {
	case hook.Unnamed:
		return nil
	case hook.Named:
		path = t.Path
	}
	transcript, err := files.Optional(path)
	if err != nil {
		return err
	}
	repo, err := git.Open(dir)
	if err != nil {
		return err
	}
	branch, err := repo.Branch()
	if err != nil {
		return err
	}
	tree, err := repo.Snapshot()
	if err != nil {
		return err
	}
	mergeBase, err := repo.Base()
	if err != nil {
		return err
	}
	diff, err := repo.Diff(mergeBase, tree)
	if err != nil {
		return err
	}
	var tokens usage.Tokens
	err = files.Update(filepath.Join(repo.GitDir, "shampoo", "transcripts.json"), func(data []byte) ([]byte, error) {
		read, err := usage.ParseOffsets(data)
		if err != nil {
			return nil, err
		}
		tokens, read[path] = usage.Read(transcript, read[path])
		return json.Marshal(read)
	})
	if err != nil {
		return err
	}
	defaultBranch, err := repo.DefaultBranch()
	if err != nil {
		return err
	}
	return files.Update(filepath.Join(repo.CommonDir, "shampoo", "ledger.jsonl"), func(data []byte) ([]byte, error) {
		for _, line := range budget.Used(rand.Text(), branch, defaultBranch, tokens, diff.Added+diff.Removed) {
			encoded, err := json.Marshal(line)
			if err != nil {
				return nil, err
			}
			data = append(append(data, encoded...), '\n')
		}
		return data, nil
	})
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
	switch earlier := judgment.ParseAttempt(data, tree).(type) {
	case judgment.Cached:
		return earlier.Verdict, nil
	case judgment.Fresh:
	}
	return "", nil
}
