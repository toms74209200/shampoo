package run

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/toms74209200/shampoo/internal/files"
	"github.com/toms74209200/shampoo/internal/git"
	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/rule"
	"github.com/toms74209200/shampoo/internal/settings"
	"github.com/toms74209200/shampoo/internal/shell"
	"github.com/toms74209200/shampoo/internal/task"
)

func Run(dir string) (rule.Outcome, string, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	data, err := files.Optional(filepath.Join(repo.Root, ".shampoo", "settings.toml"))
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	options, err := settings.Parse(data)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	if options.Tasks == "" {
		return rule.WorkflowFailure, "", errors.New("run: no task source; set [orchestrator] tasks in .shampoo/settings.toml")
	}
	code, out, err := shell.Run(repo.Root, options.Tasks)
	if err != nil || code != 0 {
		return rule.WorkflowFailure, "", fmt.Errorf("run: task source %q exited %d: %v\n%s", options.Tasks, code, err, out)
	}
	tasks, err := task.Parse(out)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	if len(tasks) == 0 {
		return rule.Pass, "no task to run\n", nil
	}
	next := tasks[0]
	worktree, err := repo.AddWorktree(next)
	if err != nil {
		return rule.WorkflowFailure, "", err
	}
	code, out, err = shell.Run(worktree, options.Agent+` "$@"`, next.DoD)
	if err != nil {
		return rule.WorkflowFailure, "", fmt.Errorf("run: agent %q: %w", options.Agent, err)
	}
	report := fmt.Sprintf("task %s: the agent exited %d\n%s", next.ID, code, out)
	outcome, guarded, err := guard.Run(worktree, rule.Task)
	return outcome, report + guarded, err
}
