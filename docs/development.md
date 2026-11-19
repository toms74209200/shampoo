# shampoo Development Guide

How shampoo itself is developed. What shampoo is and how it works is in the [design doc](design.md).

shampoo is in bootstrapping (design doc §13): it is built with a human in the loop until it can run HOOTL. This guide records the rules so that they can be enforced by shampoo once it exists, rather than living in anyone's memory. Rules are kept here and in executable checks; CLAUDE.md, agent memory, and Claude Code settings are not used to carry them.

## 1. Setup
The devcontainer runs `make setup` on creation, which enables the git hooks in `.githooks`. Outside the devcontainer, run it once by hand.

| Command | What it does |
|---|---|
| `make check` | Every check guarding a commit: lint and small tests |
| `make lint` | golangci-lint, including format checks |
| `make fmt` | Apply formatters |
| `make test-small` | Small tests (`-tags=small`) |
| `make test-acceptance` | Acceptance tests in `tests/acceptance` (pytest-bdd, run with uv) |
| `make build` | Build the binary |

golangci-lint is pinned in the Makefile and run through `go run`, so local runs and CI use the same version regardless of what the environment has installed.

## 2. Checks
| Hook | Runs |
|---|---|
| `pre-commit` | `make check` |
| `commit-msg` | `scripts/check-commit.sh --msg-file` |

Checks report by exit code: 0 pass, 1 task failure (the work is wrong), anything else a workflow failure (the check itself misbehaved), following design doc §3. `scripts/check-commit.sh` exits 1 on a rule violation and 2 on its own errors.

Lint findings have no warnings: a finding fails the check. A false positive passes only through `//nolint:<linter> // <reason>`; `nolintlint` rejects a directive without a specific linter or a reason, so exceptions stay countable from the code.

## 3. Commits
### 3.1 Minimal commits
A task (a PR) is made of minimal commits. "Minimal" is defined by what can be tried, not by judgment:

> One commit, one subject. Everything staged is needed either for the subject to be true or for the checks to pass.
>
> To try it: take out any file or hunk. If the subject is still true and the checks still pass, that part belongs in another commit.

This gives two bounds:
- **Lower bound: every commit passes the checks on its own.** A change and the white-box tests it needs or breaks are one commit, since either alone fails. The pre-commit hook checks this bound
- **Upper bound: one subject per commit.** The subject's conjunction rule (§3.2) catches the most common sign of two changes in one commit; the rest is judgment, an open problem for HOOTL (§6)

Examples from this repository's history:
- Creating the Go project is `go.mod` and `main.go` only. Dependencies, `.gitignore`, the Makefile, and the lint config each come in their own commit
- A Makefile target and the devcontainer change that calls it are two commits, the target first, so each commit works on its own
- Acceptance tests and feature specs never share a commit with the implementation they judge (design doc §5.1)

Which kinds of files a commit touches says nothing about minimality, so it is not checked.

### 3.2 Message format
```
:<gitmoji>: <Verb> <what>

<optional body: why, not what>

<optional trailers>
```

- The subject is at most 72 characters and does not end with a period
- The subject describes one change: outside `code spans` it contains no `and`, `,` or `&`. Two changes in one subject mean two commits. A single change that reads naturally with them is rephrased: `drag-and-drop` passes, as does an identifier such as `` `andThen` ``
- The message has no directive that skips CI: `[skip ci]`, `[ci skip]`, `[no ci]`, `[skip actions]`, `[actions skip]`, or a `skip-checks: true` trailer. A commit that skips CI is never checked, and the PR view cannot tell it from one that needs no checks
- A body, if any, is separated from the subject by a blank line, and explains why the change is made; the diff already shows what changed
- Emoji are written as `:code:`, not as Unicode characters
- These rules apply to commits written by people and agents. Merge commits are written by GitHub from the PR title and are not checked

`scripts/check-commit.sh` enforces this format and holds the authoritative lists of emoji and verbs. The tables below explain how to choose; change both together.

### 3.3 Emoji
| Emoji | Use for |
|---|---|
| `:sparkles:` | New feature |
| `:bug:` | Bug fix |
| `:recycle:` | Refactoring |
| `:art:` | Code structure or format |
| `:white_check_mark:` | Tests |
| `:memo:` | Documentation |
| `:hammer:` | Development scripts and tooling (Makefile, hooks, devcontainer) |
| `:construction_worker:` | CI |
| `:rotating_light:` | Linter config or lint fixes |
| `:see_no_evil:` | `.gitignore` |
| `:wrench:` | Configuration |
| `:heavy_plus_sign:` | Add a dependency |
| `:heavy_minus_sign:` | Remove a dependency |
| `:arrow_up:` | Upgrade a dependency |
| `:fire:` | Remove code or files |
| `:truck:` | Move or rename |
| `:zap:` | Performance |
| `:pencil2:` | Typo |
| `:rewind:` | Revert |

### 3.4 Verbs
The subject starts with one of the verbs below, in the imperative. Other verbs are vaguer forms of these, so the list is kept small to keep the choice predictable.

| Verb | Use for | Instead of |
|---|---|---|
| Add | Something that did not exist: a feature, file, test, target, section | Implement, Introduce, Support, Enable, Create |
| Change | Different behavior or content of something that exists | Modify, Adjust, Disable, Use |
| Update | Bringing something up to date with something else, such as docs to match code | Upgrade, Pin, Sync |
| Bump | Raising a dependency's version. Lowercase `bump` is also allowed, as Dependabot writes it after a prefix | |
| Fix | Correcting wrong behavior | Correct, Resolve |
| Remove | Deleting something | Delete, Drop, Disable |
| Rename | A new name, nothing else | |
| Move | A new location, nothing else | |
| Refactor | New structure, same behavior | Split, Extract, Simplify, Improve, Reorganize |
| Clean | Tidying without changing structure or behavior | Tidy |
| Replace | Swapping one thing for another | Use, Switch |
| Apply | The mechanical output of a tool, such as a formatter | |
| Revert | Undoing an earlier commit | |

## 4. Branches and merging
- **Now (bootstrapping):** commits go directly to `master`, each one minimal
- **Repository rules:** the `master` ruleset requires a PR and merge commits, and forbids force pushes and deletion. Auto-merge is enabled, and only merge commits are allowed, titled with the PR title. Until production code starts, repository admins bypass the ruleset so that the commits above can be pushed directly. **Remove that bypass before the first production code**
- **Production code:** one task per branch and PR, made of minimal commits. The PR is merged with a merge commit once the checks its changes call for pass, so each task's commits stay on `master` as they are

## 5. CI
Each workflow starts from the changes it checks, as in earlier repositories:

| Workflow | Starts when | Checks |
|---|---|---|
| `on-pullrequest-commit.yml` | Every PR, since every PR has commits | The message of each commit |
| `on-pullrequest-go.yml` | Go-related files change (its `paths`) | Applies the formatter and commits the result, then lint; small, medium and acceptance tests; coverage |
| `on-pullrequest-python.yml` | Acceptance tests change | Acceptance tests |
| `on-pullrequest-features.yml` | Feature specs change | Scenario annotations, statically (design doc §5.1) |

The `job-*.yml` workflows hold the steps, and the `on-*.yml` workflows decide when they run.

Rules for CI:
- **A run's results belong to the commit that started it.** A run may change the code before checking it (apply a formatter, generate, build, bundle) and may push what it produced; those commits are its output, covered by its results
- **Changes made by automation arrive as PRs.** Whatever makes them, an agent, a scheduled workflow or a bot, opens its own PR rather than pushing to another PR's branch, and that PR runs CI like any other. A PR opened with the workflow token starts no workflows, so automation uses a token that does (e.g. a GitHub App's)
- **Nothing skips CI.** Commit messages cannot skip it (§3.2)
- **Tool versions are pinned in the Makefile, and CI runs `make`**, so local runs and CI judge the same way

No check is required by the ruleset. GitHub's required checks are a fixed list on the head commit, while the checks a PR needs depend on its changes, and the commits CI pushes run no checks of their own. A human merges until shampoo does; shampoo then decides from the changed paths which checks must have passed, on the commit that started them (design doc §10).

Dependabot, when added, uses the commit prefix `:arrow_up:` so that its commits follow the commit rules, with `bump` as an allowed verb. The exact message it produces is to be verified then.

## 6. Open problems
Recorded here so that they are solved before shampoo runs HOOTL:
- **The upper bound of a minimal commit** (§3.1) is kept by judgment, which needs a human or an agent's discretion
