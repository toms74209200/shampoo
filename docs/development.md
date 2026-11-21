# shampoo Development Guide

How shampoo itself is developed. What shampoo is and how it works is in the [design doc](design.md).

shampoo is in bootstrapping (design doc §13): it is built with a human in the loop until it can run HOOTL. This guide records the rules so that they can be enforced by shampoo once it exists, rather than living in anyone's memory. Rules are kept here and in executable checks; CLAUDE.md, agent memory, and Claude Code settings are not used to carry them. Claude Code settings hold only the hooks that run shampoo, the permissions that deny commands the agent must not run, and `"autoMemoryEnabled": false`, so that rules cannot drift into agent memory.

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

golangci-lint is pinned in `scripts/golangci-lint.sh`, which installs that version, so local runs and CI use the same version regardless of what the environment has installed.

## 2. Checks
| Hook | Runs |
|---|---|
| `pre-commit` | `make check` on what is staged, in a clean worktree |
| `commit-msg` | `scripts/check-commit.sh --msg-file` |
| Claude Code `PostToolUse` on every tool, and `Stop` (`.claude/settings.json`) | `scripts/hook.sh guard edit` and `scripts/hook.sh guard task`: builds shampoo from the working tree, falling back to the last build when it does not build, and runs `shampoo guard edit` or `shampoo guard task`. A task failure or a budget exceeded, which calls for harness repair, is sent back to the agent (exit 2); a workflow failure or repairs beyond their budget are shown to the human (exit 1) |
| Claude Code `SessionStart` on `startup`, `clear` and `compact` (`.claude/settings.json`) | `scripts/hook.sh policies`: builds shampoo the same way and prints every policy with `shampoo policies`; the output enters the agent's context (design doc §5.2) |

Claude Code permissions in `.claude/settings.json` deny making a branch with `git checkout -b`; branches are made with `git switch -c`.

Checks report by exit code: 0 pass, 1 task failure (the work is wrong), anything else a workflow failure (the check itself misbehaved), following design doc §3. `scripts/check-commit.sh` exits 1 on a rule violation and 2 on its own errors. A check's command calls a script or tool directly, not a make target: make exits 2 whenever a recipe fails, which would turn every task failure into a workflow failure.

Lint findings have no warnings: a finding fails the check. A sum type is a sealed interface marked `//sumtype:decl`, and `gochecksumtype` fails a type switch over it that misses a case, the check Go's compiler does not make (policy sum-types). The side-effect packages (`scripts/side-effect-packages.txt`) are statements only: `scripts/check-side-effect-statements.sh` fails any choice in them other than handling an error or matching a sum type, since what they do is the libraries' and has nothing of ours to test. Choices belong in pure code, whose share and small-test coverage the coverage job reports. A false positive passes only through `//nolint:<linter> // <reason>`; `nolintlint` rejects a directive without a specific linter or a reason, so exceptions stay countable from the code. `scripts/lint.sh` writes them to `out/exceptions/lint.txt`, which the lint rule declares as its exceptions, at most 3 (design doc §3.1).

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
- **Tasks:** one task per branch and PR, made of minimal commits. Branches are made with `git switch -c`
- **Repository rules:** the `master` ruleset requires a PR and merge commits, and forbids force pushes and deletion. No check is required (§5). Repository admins still bypass the ruleset, and GitHub's auto-merge is still enabled though not used. **Remove that bypass and turn auto-merge off**: production code has started, and only shampoo merges (design doc §10)
- **Merging:** the PR is merged with a merge commit, titled with the PR title, once the checks its changes call for pass, so each task's commits stay on `master` as they are. A human merges until `shampoo next` does (design doc §10)

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
- **A run that pushes verifies what it pushed.** A run may change the code before checking it (apply a formatter, generate, build, bundle) and may push what it produced; it then checks the pushed code in the same run, after the push. The runs GitHub creates on such a commit do not execute and are not results
- **Changes made by automation arrive as PRs.** Whatever makes them, an agent, a scheduled workflow or a bot, opens its own PR rather than pushing to another PR's branch, and that PR runs CI like any other. A PR opened with the workflow token starts no workflows, so automation uses a token that does (e.g. a GitHub App's)
- **Nothing skips CI.** Commit messages cannot skip it (§3.2)
- **Tool versions are pinned in the scripts that run them, and CI runs `make`**, whose targets call those scripts, so local runs and CI judge the same way

No check is required by the ruleset. GitHub's required checks are a fixed list on the head commit, while the checks a PR needs depend on its changes, and the commits CI pushes run no checks of their own. shampoo instead decides from the changed paths which workflows must have passed, counting the latest run of each on any of the PR's commits (design doc §10).

Dependabot, when added, uses the commit prefix `:arrow_up:` so that its commits follow the commit rules, with `bump` as an allowed verb. The exact message it produces is to be verified then.

## 6. Open problems
Recorded here so that they are solved before shampoo runs HOOTL:
- **The upper bound of a minimal commit** (§3.1) is kept by judgment, which needs a human or an agent's discretion
