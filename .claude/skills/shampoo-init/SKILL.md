---
name: shampoo-init
description: Set up shampoo in a repository up to a passing guard. Use when asked to adopt, install, or set up shampoo in a project, alone or from a reference repository.
---

# Setting up shampoo

Set up shampoo in the target repository. Decide what the target and any
reference repository already settle, ask only what remains open, and finish
with a passing guard.

## 1. Inspect the repository

Check the target repository first, then any reference repository:

- Linter, its config, and how it is invoked
- How tests run, and which are small (no network, no files outside a temp dir, no external processes)
- Coverage command and threshold
- `.github/workflows/` and how pull requests merge
- Commit message conventions (`git log`, commit checks)

If the reference repository has `.shampoo/`, copy and adapt its rules, checks,
and policies first. `shampoo init` keeps every existing file.

## 2. Ask only what is open

Decide everything the repositories already settle. Ask the human only where
they leave a choice open, or where setup would change the project's CI
(including `.github/workflows/`) or how pull requests merge. Ask all such
questions at once, each with your proposal, so setup does not become a
dialogue. When asked for hands-off setup, ask nothing and record your decisions
in the pull request description.

## 3. Assemble check commands

Provide a command for each recommended rule the project can support (omit the
rest):

| Rule | Command |
|---|---|
| `lint` | Runs the linter; exits `1` on violation. If inline exceptions are allowed, write one line per exception to a file and set `exceptions` |
| `small-tests` | Runs small tests; exits `1` on failure |
| `coverage` | Checks small-test coverage; exits `1` below threshold (`80%` is a start) |
| `commit-message` | Checks commit `<commit>`; exits `1` on format violation |

Commands must exit `0` on pass, `1` on a task violation, and any other non-zero
code when the check itself fails. Call tools or `scripts/` directly rather than
`make` (which exits `2` on recipe failure). Add `paths` (e.g.
`["**.go", "go.mod"]`) or `size` (`"small"`, `"medium"`, `"large"`) as needed.

## 4. Add initial tasks

Unless tasks with completion criteria already exist elsewhere, create `tasks/`
with a few small one-PR tasks (smallest first):

```markdown
# <what to do>

## Done when
- <what a test can check>
```

## 5. Run `shampoo init`

```sh
echo '{"tasks": "<task source>", "checks": {"lint": {"command": "...", "exceptions": "..."}, "small-tests": {"command": "..."}, "coverage": {"command": "..."}, "commit-message": {"command": "... <commit>"}}}' | shampoo init
```

Writes `.shampoo/`, recommended policies, and hooks in `.claude/settings.json`
without overwriting existing files or settings. If `stderr` reports no review
phase, keep it so only if the project wants no policy review.

## 6. Verify and open a PR

Run each check command and `shampoo guard task`; fix any check that fails on
clean code. Commit the setup on its own branch and open a pull request with
what you decided.
