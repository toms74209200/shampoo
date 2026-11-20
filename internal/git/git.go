package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toms74209200/shampoo/internal/rule"
)

type Repository struct {
	Root      string
	CommonDir string
	GitDir    string
}

func Open(dir string) (Repository, error) {
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--show-toplevel", "--absolute-git-dir", "--git-common-dir")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return Repository{}, fmt.Errorf("git rev-parse: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 3 {
		return Repository{}, fmt.Errorf("git rev-parse: unexpected output %q", out)
	}
	return Repository{Root: lines[0], GitDir: lines[1], CommonDir: lines[2]}, nil
}

func (r Repository) Snapshot() (string, error) {
	env := append(os.Environ(), "GIT_INDEX_FILE="+filepath.Join(r.GitDir, "shampoo-index"))
	add := exec.Command("git", "add", "--all")
	add.Dir = r.Root
	add.Env = env
	if out, err := add.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %w: %s", err, out)
	}
	write := exec.Command("git", "write-tree")
	write.Dir = r.Root
	write.Env = env
	out, err := write.Output()
	if err != nil {
		return "", fmt.Errorf("git write-tree: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r Repository) Passed(transition rule.Transition) (string, error) {
	recorded := exec.Command("git", "rev-parse", "--quiet", "--verify", "refs/worktree/shampoo/"+string(transition))
	recorded.Dir = r.Root
	if out, err := recorded.Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	} else if exit := (*exec.ExitError)(nil); !errors.As(err, &exit) {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	head := exec.Command("git", "rev-parse", "--verify", "HEAD^{tree}")
	head.Dir = r.Root
	out, err := head.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r Repository) Changed(from, to string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "-z", "--no-renames", from, to)
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	var changed []string
	for _, path := range bytes.Split(out, []byte{0}) {
		if len(path) > 0 {
			changed = append(changed, string(path))
		}
	}
	return changed, nil
}

func (r Repository) Record(transition rule.Transition, tree string) error {
	cmd := exec.Command("git", "update-ref", "refs/worktree/shampoo/"+string(transition), tree)
	cmd.Dir = r.Root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git update-ref: %w: %s", err, out)
	}
	return nil
}

func (r Repository) Branch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r Repository) Unmerged() ([]string, error) {
	cmd := exec.Command("git", "branch", "--no-merged", "HEAD", "--format=%(refname:short)")
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch --no-merged: %w", err)
	}
	return strings.Fields(string(out)), nil
}

func (r Repository) DefaultBranch() (string, error) {
	for _, ref := range []string{"refs/remotes/origin/HEAD", "refs/heads/main", "refs/heads/master"} {
		cmd := exec.Command("git", "rev-parse", "--quiet", "--verify", "--symbolic-full-name", ref)
		cmd.Dir = r.Root
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(out))
		name = strings.TrimPrefix(strings.TrimPrefix(name, "refs/remotes/origin/"), "refs/heads/")
		return name, nil
	}
	return "", errors.New("git: no default branch among origin/HEAD, main and master")
}

func (r Repository) Base() (string, error) {
	branch, err := r.DefaultBranch()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "merge-base", "HEAD", branch)
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git merge-base %s: %w", branch, err)
	}
	return strings.TrimSpace(string(out)) + "^{tree}", nil
}
