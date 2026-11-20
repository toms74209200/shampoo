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
	if options.Tasks == "" {
		return next.Action{}, errors.New("next: no task source; set [orchestrator] tasks in .shampoo/settings.toml")
	}
	if !strings.Contains(options.Tasks, "://") {
		if _, err := os.Stat(filepath.Join(repo.Root, options.Tasks)); err != nil {
			return next.Action{}, fmt.Errorf("next: task source: %w", err)
		}
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
	state := next.State{Source: options.Tasks, Branch: branch, Default: branch == defaultBranch, Verdict: verdict}
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
	if state.Default {
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
		state.MaxOpen = options.MaxOpen
		return next.Decide(state), nil
	}
	policies, err := Policies(repo.Root)
	if err != nil {
		return next.Action{}, err
	}
	for _, p := range policies {
		if p.Review {
			state.Policies = append(state.Policies, p.ID)
		}
	}
	state.Requirements = options.Requirements
	recorded, err := files.Optional(filepath.Join(repo.GitDir, "shampoo", "review.json"))
	if err != nil {
		return next.Action{}, err
	}
	if recorded != nil {
		var last reviewed
		if err := json.Unmarshal(recorded, &last); err != nil {
			return next.Action{}, fmt.Errorf("next: review: %w", err)
		}
		tree, err := repo.Snapshot()
		if err != nil {
			return next.Action{}, err
		}
		state.Reviewed = true
		if last.Tree == tree {
			state.Findings = last.Findings
		}
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
		}
		state.PullRequest.Pushed = slices.Contains(state.PullRequest.Commits, strings.TrimSpace(string(out)))
		var changed []string
		for _, file := range pr.Files {
			changed = append(changed, file.Path)
		}
		for _, suffix := range []string{".yml", ".yaml"} {
			found, err := files.Read(filepath.Join(repo.Root, ".github", "workflows"), suffix)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return next.Action{}, err
			}
			for _, file := range found {
				parsed, err := workflow.Parse(".github/workflows/"+file.Name+suffix, file.Data)
				if err != nil {
					return next.Action{}, err
				}
				if parsed.Expects(pr.BaseRefName, changed) {
					state.PullRequest.Expected = append(state.PullRequest.Expected, parsed.Name)
				}
			}
		}
		list, err := gh("run", "list", "--branch", branch, "--limit", "200", "--json", "workflowName,headSha,status,conclusion,url,createdAt")
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
	action := next.Decide(state)
	switch action.Action {
	case "merge":
		if _, err := gh("pr", "merge", strconv.Itoa(pr.Number), "--"+options.Merge); err != nil {
			return next.Action{}, fmt.Errorf("next: %w", err)
		}
		action.Action = "merged"
	case "wait":
		latest := pr.Commits[len(pr.Commits)-1].CommittedDate
		if limit := time.Duration(options.MaxWaitMinutes) * time.Minute; time.Since(latest) >= limit {
			return next.Action{}, fmt.Errorf("next: %s has waited for its checks since %s, longer than [orchestrator] max-wait-minutes (%d)", branch, latest.Format(time.RFC3339), options.MaxWaitMinutes)
		}
	}
	return action, nil
}

type reviewed struct {
	Tree     string           `json:"tree"`
	Findings []review.Finding `json:"findings"`
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
	encoded, err := json.Marshal(reviewed{Tree: tree, Findings: findings})
	if err != nil {
		return err
	}
	if err := files.Update(filepath.Join(repo.GitDir, "shampoo", "review.json"), func([]byte) ([]byte, error) { return encoded, nil }); err != nil {
		return err
	}
	run := rand.Text()
	lines := []budget.Line{{Run: run, Task: branch, Transition: "review", Outcome: rule.Pass}}
	if len(findings) > 0 {
		lines = nil
	}
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
