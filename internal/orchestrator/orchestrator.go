package orchestrator

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/toms74209200/shampoo/internal/budget"
	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/next"
	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/question"
	"github.com/toms74209200/shampoo/internal/review"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
	"github.com/toms74209200/shampoo/internal/workflow"
)

func Next(dir string) (next.Action, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return next.Action{}, err
	}
	data, err := files.Optional(filepath.Join(repo.Root, ".shampoo", "settings.toml"))
	if err != nil {
		return next.Action{}, err
	}
	options, err := settings.Parse(data)
	if err != nil {
		return next.Action{}, err
	}
	branch, err := repo.Branch()
	if err != nil {
		return next.Action{}, err
	}
	defaultBranch, err := repo.DefaultBranch()
	if err != nil {
		return next.Action{}, err
	}
	verdict, err := guard.Last(repo.Root, rule.Task)
	if err != nil {
		return next.Action{}, err
	}
	_, missing := os.Stat(filepath.Join(repo.Root, options.Tasks))
	state := next.Observation{Source: options.Tasks, SourceFound: missing == nil, Branch: branch, Default: branch == defaultBranch, Verdict: verdict, MaxOpen: options.MaxOpen, Requirements: options.Requirements}
	gh := func(args ...string) ([]byte, error) {
		cmd := exec.Command("gh", args...)
		cmd.Dir = repo.Root
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("gh %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
		}
		return out, nil
	}
	list, err := gh("issue", "list", "--label", options.QuestionLabel, "--state", "open", "--limit", "1000", "--json", "url")
	if err != nil {
		return next.Action{}, fmt.Errorf("next: %w", err)
	}
	var questions []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(list, &questions); err != nil {
		return next.Action{}, fmt.Errorf("next: gh issue list: %w", err)
	}
	for _, question := range questions {
		state.Questions = append(state.Questions, question.URL)
	}
	list, err = gh("issue", "list", "--label", options.WorkflowReview.Label, "--state", "open", "--limit", "1000", "--json", "url")
	if err != nil {
		return next.Action{}, fmt.Errorf("next: %w", err)
	}
	var escalations []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(list, &escalations); err != nil {
		return next.Action{}, fmt.Errorf("next: gh issue list: %w", err)
	}
	for _, escalation := range escalations {
		state.Escalations = append(state.Escalations, escalation.URL)
	}
	statsPath := filepath.Join(repo.CommonDir, "shampoo", "stats.jsonl")
	stats, err := files.Optional(statsPath)
	if err != nil {
		return next.Action{}, err
	}
	state.Due = budget.Reviewing(budget.ParseStats(stats), options.WorkflowReview)
	state.Policies, err = Policies(repo.Root)
	if err != nil {
		return next.Action{}, err
	}
	recorded, err := files.Optional(filepath.Join(repo.GitDir, "shampoo", "review.json"))
	if err != nil {
		return next.Action{}, err
	}
	state.Record, err = next.ParseRecord(recorded)
	if err != nil {
		return next.Action{}, fmt.Errorf("next: %w", err)
	}
	state.Tree, err = repo.Snapshot()
	if err != nil {
		return next.Action{}, err
	}
	var pr struct {
		Number      int    `json:"number"`
		State       string `json:"state"`
		BaseRefName string `json:"baseRefName"`
		Commits     []struct {
			Oid           string    `json:"oid"`
			CommittedDate time.Time `json:"committedDate"`
		} `json:"commits"`
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	view, err := gh("pr", "view", branch, "--json", "number,state,baseRefName,commits,files")
	if err == nil {
		if err := json.Unmarshal(view, &pr); err != nil {
			return next.Action{}, fmt.Errorf("next: gh pr view: %w", err)
		}
		head := exec.Command("git", "rev-parse", "HEAD")
		head.Dir = repo.Root
		out, err := head.Output()
		if err != nil {
			return next.Action{}, fmt.Errorf("git rev-parse HEAD: %w", err)
		}
		state.PullRequest = &next.PullRequest{State: pr.State}
		for _, commit := range pr.Commits {
			state.PullRequest.Commits = append(state.PullRequest.Commits, commit.Oid)
			state.PullRequest.Latest = commit.CommittedDate
		}
		state.PullRequest.Pushed = slices.Contains(state.PullRequest.Commits, strings.TrimSpace(string(out)))
		var changed []string
		for _, file := range pr.Files {
			changed = append(changed, file.Path)
		}
		var sources []workflow.Source
		for _, suffix := range []string{".yml", ".yaml"} {
			found, err := files.Read(filepath.Join(repo.Root, ".github", "workflows"), suffix)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return next.Action{}, err
			}
			for _, file := range found {
				sources = append(sources, workflow.Source{Path: ".github/workflows/" + file.Name + suffix, Data: file.Data})
			}
		}
		state.PullRequest.Expected, err = workflow.Expected(sources, pr.BaseRefName, changed)
		if err != nil {
			return next.Action{}, err
		}
		list, err = gh("run", "list", "--branch", branch, "--limit", "200", "--json", "workflowName,headSha,status,conclusion,url,createdAt")
		if err != nil {
			return next.Action{}, fmt.Errorf("next: %w", err)
		}
		var runs []struct {
			WorkflowName string    `json:"workflowName"`
			HeadSha      string    `json:"headSha"`
			Status       string    `json:"status"`
			Conclusion   string    `json:"conclusion"`
			URL          string    `json:"url"`
			CreatedAt    time.Time `json:"createdAt"`
		}
		if err := json.Unmarshal(list, &runs); err != nil {
			return next.Action{}, fmt.Errorf("next: gh run list: %w", err)
		}
		for _, run := range runs {
			state.PullRequest.Runs = append(state.PullRequest.Runs, next.Run{Workflow: run.WorkflowName, Commit: run.HeadSha, Status: run.Status, Conclusion: run.Conclusion, URL: run.URL, Created: run.CreatedAt})
		}
	}
	state.Now = time.Now()
	state.MaxWait = time.Duration(options.MaxWaitMinutes) * time.Minute
	parsed := next.Parse(state)
	switch s := parsed.(type) {
	case next.Passed:
		if _, err := gh("pr", "merge", strconv.Itoa(pr.Number), "--"+options.Merge); err != nil {
			return next.Action{}, fmt.Errorf("next: %w", err)
		}
		parsed = next.Merged(s)
	case next.ReviewingWorkflow:
		encoded, err := json.Marshal(budget.StatLine{Review: "workflow"})
		if err != nil {
			return next.Action{}, err
		}
		if err := files.Update(statsPath, func(data []byte) ([]byte, error) { return append(append(data, encoded...), '\n'), nil }); err != nil {
			return next.Action{}, err
		}
	case next.Escalating:
		created, err := gh("issue", "create", "--title", s.Title, "--label", options.WorkflowReview.Label, "--body", s.Body)
		if err != nil {
			return next.Action{}, fmt.Errorf("next: %w", err)
		}
		issue := strings.TrimSpace(string(created))
		encoded, err := json.Marshal(budget.StatLine{Review: "human", Issue: issue})
		if err != nil {
			return next.Action{}, err
		}
		if err := files.Update(statsPath, func(data []byte) ([]byte, error) { return append(append(data, encoded...), '\n'), nil }); err != nil {
			return next.Action{}, err
		}
		parsed = next.Blocked{Questions: []string{issue}}
	case next.NoSource:
		return next.Action{}, fmt.Errorf("next: the task source %q is not set or not found; set [orchestrator] tasks in .shampoo/settings.toml", s.Source)
	case next.Stalled:
		return next.Action{}, fmt.Errorf("next: %s has waited for its checks since %s, longer than [orchestrator] max-wait-minutes (%d)", s.Branch, s.Since.Format(time.RFC3339), int(s.MaxWait.Minutes()))
	default:
	}
	return next.Decide(parsed), nil
}

func Policies(dir string) ([]policy.Policy, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return nil, err
	}
	sources, err := files.Read(filepath.Join(repo.Root, ".shampoo", "policies"), ".toml")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	var policies []policy.Policy
	for _, source := range sources {
		parsed, err := policy.Parse(source.Name, source.Data)
		if err != nil {
			return nil, err
		}
		policies = append(policies, parsed)
	}
	return policies, nil
}

func Ask(dir string, input io.Reader) (string, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return "", err
	}
	data, err := files.Optional(filepath.Join(repo.Root, ".shampoo", "settings.toml"))
	if err != nil {
		return "", err
	}
	options, err := settings.Parse(data)
	if err != nil {
		return "", err
	}
	branch, err := repo.Branch()
	if err != nil {
		return "", err
	}
	asked, err := io.ReadAll(input)
	if err != nil {
		return "", err
	}
	issue, err := question.Parse(asked, branch)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("gh", "issue", "create", "--title", issue.Title, "--label", options.QuestionLabel, "--body", issue.Body)
	cmd.Dir = repo.Root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh issue create: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	encoded, err := json.Marshal(budget.Line{Run: rand.Text(), Task: branch, Transition: "ask", Outcome: rule.Question})
	if err != nil {
		return "", err
	}
	err = files.Update(filepath.Join(repo.CommonDir, "shampoo", "ledger.jsonl"), func(ledger []byte) ([]byte, error) {
		return append(append(ledger, encoded...), '\n'), nil
	})
	return strings.TrimSpace(string(out)), err
}

func Review(dir string, input io.Reader) error {
	repo, err := git.Open(dir)
	if err != nil {
		return err
	}
	policies, err := Policies(repo.Root)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	findings, err := review.Parse(data, policies)
	if err != nil {
		return err
	}
	tree, err := repo.Snapshot()
	if err != nil {
		return err
	}
	branch, err := repo.Branch()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(next.Record{Branch: branch, Tree: tree, Findings: findings})
	if err != nil {
		return err
	}
	if err := files.Update(filepath.Join(repo.GitDir, "shampoo", "review.json"), func([]byte) ([]byte, error) { return encoded, nil }); err != nil {
		return err
	}
	run := rand.Text()
	lines := []budget.Line{{Run: run, Task: branch, Transition: "review", Outcome: rule.Pass}}
	for _, finding := range findings {
		lines = append(lines, budget.Line{Run: run, Task: branch, Transition: "review", Policy: finding.Policy, Outcome: rule.Finding})
	}
	return files.Update(filepath.Join(repo.CommonDir, "shampoo", "ledger.jsonl"), func(ledger []byte) ([]byte, error) {
		for _, line := range lines {
			encoded, err := json.Marshal(line)
			if err != nil {
				return nil, err
			}
			ledger = append(append(ledger, encoded...), '\n')
		}
		return ledger, nil
	})
}
