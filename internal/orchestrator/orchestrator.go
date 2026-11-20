package orchestrator

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/next"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
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
	if branch != defaultBranch && verdict == rule.Pass {
		view := exec.Command("gh", "pr", "view", branch, "--json", "state")
		view.Dir = repo.Root
		if view.Run() == nil {
			return next.Action{}, fmt.Errorf("next: %s has a pull request; following its checks is not implemented yet", branch)
		}
	}
	return next.Decide(next.State{Source: options.Tasks, Branch: branch, Default: branch == defaultBranch, Verdict: verdict}), nil
}
