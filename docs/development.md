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
- **Upper bound: one subject per commit.** The subject's conjunction rule (§3.2) catches the most common sign of two changes in one commit; the rest is judgment, an open problem for HOOTL (§8)

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

## 4. Code
Rules for production code. Each rule names its deterministic check; a check marked *(not yet)* is not implemented. What no check can decide is left to review.

### 4.1 Comments
Code has no comments, test code included. What the code does is stated by the code and its types; why it is so goes in the commit message body or the design doc; how to work on the project goes in this guide. Only directives are allowed: `//go:` directives, build constraints, `// Code generated`, and `//nolint:<linter> // <reason>` (§4.6).

Check *(not yet)*: any other comment fails.

### 4.2 Dependencies
Dependencies are kept to the minimum. A module is added only with a reason, recorded in an allowlist. The allowlist is the project's policy: another organization replaces it with its own rules, and the check stays the same. Changing it changes what judges the work, so it goes through human review.

Tools such as golangci-lint are not dependencies: they are pinned in the Makefile and run through `go run`. Dependencies used only by acceptance tests live in the acceptance tests' own module.

Check *(not yet)*: every module in `go list -m all` is in the allowlist.

### 4.3 Side effects
Side effects (the file system, network, processes, databases, the clock, randomness, environment variables) happen only in the packages designated for them. Everything else is pure: inputs are arguments, outputs are return values, and nothing keeps state. This is what lets most code be covered by small tests (§5.1).

Check *(not yet)*: outside the designated packages, `depguard` rejects imports such as `os`, `os/exec`, `net`, `net/http` and `database/sql`, and `forbidigo` rejects calls such as `time.Now`, `time.Sleep`, `math/rand` and `os.Getenv`.

### 4.4 Branches go into types and structure
Every branch is something a test has to cover. Before writing a branch, move it into a type or a structure, where the compiler or the structure guarantees it and no test is needed.

**Parse, don't validate.** A function that establishes a fact returns a type that carries it. A result whose type does not carry the fact forgets it, so every caller checks it again: `string -> string`, `string -> int`, `int -> int` and `User -> User` all lose what the function found out.
Instead of:
```go
func normalizePort(s string) (int, error)
if port > 0 && port < 65536 { ... }
```
Write:
```go
type Port uint16
func ParsePort(s string) (Port, error)
```
Tests no longer needed: every caller's re-check.

**Sum types and exhaustive type switches.** Variants are types of a sealed interface, not values of a kind field.
Instead of:
```go
type Command struct{ Kind, Name string }
switch c.Kind { case "shell": ...; case "stop": ... }
```
Write:
```go
type Command interface{ command() }
type Shell struct{ Name string }
type Stop struct{ Name string }
func (Shell) command() {}
func (Stop) command()  {}
switch c := cmd.(type) { case Shell: ...; case Stop: ... }
```
`gochecksumtype` checks that every switch covers every variant *(not yet)*. Tests no longer needed: unknown kinds, and fields that one kind leaves unset.

**Typestate.** A state with different rules is a different type, not a flag.
Instead of:
```go
type User struct{ Active bool }
func Approve(u User) error { if !u.Active { return errInactive }; ... }
```
Write:
```go
type ActiveUser struct{ ... }
func Activate(u User) (ActiveUser, error)
func Approve(u ActiveUser)
```
Tests no longer needed: approving an inactive user.

**Polymorphism instead of repeated switches (Strategy, State).** When several functions switch on the same thing, each case becomes one implementation of an interface.
Instead of:
```go
func render(f Format, d Doc) string { switch f { ... } }
func extension(f Format) string    { switch f { ... } }
```
Write:
```go
type Format interface {
	Render(Doc) string
	Extension() string
}
```
Tests no longer needed: the same switch tested once per function.

**Chain of Responsibility.** A growing if-else ladder becomes a list of handlers; each one decides alone whether it handles the input.
```go
type Handler interface{ Handle(Request) (Response, bool) }
handlers := []Handler{RedirectHandler{}, StaticHandler{}, NotFoundHandler{}}
```
Each handler is tested alone, and adding one changes no other handler or test.

**Lookup tables.** When cases differ only in data, the data is a table, not branches.
Instead of:
```go
switch code { case 200: return "OK"; case 404: return "Not Found"; ... }
```
Write:
```go
var statusText = map[Status]string{OK: "OK", NotFound: "Not Found"}
```

**Split flag arguments.** A boolean argument that switches behavior becomes two functions.
Instead of:
```go
func Render(d Doc, compact bool) string
```
Write:
```go
func RenderCompact(d Doc) string
func RenderFull(d Doc) string
```

**Decorator.** Cross-cutting concerns such as retries or logging wrap the core instead of branching inside it.
Instead of:
```go
func Fetch(url string, retry bool) (Page, error) { ... if retry { ... } }
```
Write:
```go
type Fetcher interface{ Fetch(URL) (Page, error) }
type Retrying struct{ Inner Fetcher; Attempts int }
```

**Method overloading** chooses by argument types at compile time. Go has no overloading, so it does not apply here; it is listed for languages that have it.

The Null Object pattern is an anti-pattern and is not used.

### 4.5 When a function exists
A function exists only when its result type carries what it established (§4.4) and it hides branches behind a small interface (a deep module). These are not functions and are inlined:
- A bundle of statements wrapped for readability
- Statements that look the same, extracted only to share them
- A call one level down, or a thin entry point that only passes values on
- A name for a condition used in one place
- A meaningless conversion: `string -> string`, `string -> int`, `User -> User`

Splitting code into such functions makes each piece look short while the code as a whole grows, and every piece needs a test of its own.

Check *(not yet)*: functions whose body is a single expression (except one returning a function literal); unexported functions without branches and with one caller; functions whose results callers branch on; functions with only primitive parameters and results; functions returning a parameter's type. Methods whose signature an interface fixes, such as `UnmarshalJSON`, are exempt.

### 4.6 Exceptions
`//nolint:<linter> // <reason>` is the only exception, and only for a false positive. Adding one loosens a check that judges the work, so it goes through human review. The exceptions per check are counted from the code; when they exceed the check's budget, the check is fixed instead of adding more (design doc §5).

## 5. Tests
Code and tests are two sides of one thing. Tests are as many small tests as possible: fast, and covering the code.

### 5.1 Sizes
Tests follow test sizes (design doc §5), selected by build tag: `small`, `medium`, `large`.

| | Small | Medium | Large |
|---|---|---|---|
| Time limit | 60 s | 300 s | 900 s or more |
| Network | No | localhost only | Yes |
| Database | No | Yes | Yes |
| File system | No | Avoided | Yes |
| Threads | No | Yes | Yes |
| Sleep | No | Yes | Yes |
| Environment | No | Yes | Yes |

- Pure code (§4.3) is tested by small tests. They test values, which is what makes parameterized tests and property-based tests possible
- Code with side effects is tested by medium or larger tests, written knowing their cost: they are slower, and they need fakes or real resources
- A server whose code is mostly CRUD may have few small tests; that is accepted, since coverage is counted across all sizes (§5.4)

Check *(not yet)*: tests in packages without side effects carry `//go:build small`, and the import rules of §4.3 apply to them too; `-timeout` enforces the time limits.

### 5.2 What a test depends on
- **A white-box test depends on its target only.** If the target is a function, no other production code is used, except the types needed to build its arguments
- **A black-box test depends on the public API only.** It lives in an external test package (`package foo_test`), so the compiler allows nothing else
- **No mocks.** Fakes stand in for what has side effects

Check *(not yet)*: in a white-box test, production identifiers other than the target named by the test (§5.3) and types fail. No mocking library is in the dependency allowlist (§4.2).

### 5.3 How a test is written
- **Name:** `when <target> with <given conditions> then returns <expected>`
- **DAMP over DRY.** Each test reads on its own
- **No test helpers.** A test file declares only `Test`, `Fuzz` and `Example` functions, and does not call `t.Helper()`
- **No fixed strings or numbers.** Inputs are random, from the standard library (`crypto/rand.Text`, `testing/quick`), unless the value itself is the point, such as a boundary
- **Parameterized tests** when cases differ only in inputs and expected results
- **Property-based tests** when a property holds for all inputs (`testing/quick`, or fuzz targets whose seeds run as tests)
- **Arrange-Act-Assert.** Enforced at medium size and larger once tests are instrumented with OpenTelemetry, where each phase becomes a span
- **Feature specs and acceptance tests are the spec.** They change only in a spec-change commit that precedes the implementation, never together with it (design doc §5.1)

Check *(not yet)*: names of subtests and table cases against the pattern; declarations and `t.Helper()` in test files; string and number literals in tests; duplicated test bodies (`dupl`).

### 5.4 Coverage and mutation testing
- **Coverage is one ratio over the whole code base, across all sizes.** `-coverpkg=./...` counts packages without tests as uncovered, and coverage from binaries built with `-cover` and run by medium and large tests is merged with `go tool covdata merge`. No list of packages to measure is kept, so new code cannot escape it
- **Coverage does not go down** (a ratchet), rather than meeting a fixed threshold. A threshold invites tests written only to reach it
- **Mutation testing shows whether tests assert.** Coverage only shows that code ran. On a PR it runs on the changed lines; at a milestone, on everything. The tool is pinned in the Makefile like golangci-lint

Check *(not yet)*: all of the above.

## 6. Branches and merging
- **Now (bootstrapping):** commits go directly to `master`, each one minimal
- **Repository rules:** the `master` ruleset requires a PR, the commit message check (`commit-message / check`), and merge commits, and forbids force pushes and deletion. Auto-merge is enabled, and only merge commits are allowed, titled with the PR title. Until production code starts, repository admins bypass the ruleset so that the commits above can be pushed directly. **Remove that bypass before the first production code**
- **Production code:** one task per branch and PR, made of minimal commits. The PR is merged with a merge commit once the checks its changes call for pass, so each task's commits stay on `master` as they are

## 7. CI
Each workflow starts from the changes it checks, as in earlier repositories:

| Workflow | Starts when | Checks |
|---|---|---|
| `on-pullrequest-commit.yml` | Every PR, since every PR has commits | The message of each commit |
| `on-pullrequest-go.yml` | Go-related files change (its `paths`) | Applies the formatter and commits the result, then lint; small tests |

The `job-*.yml` workflows hold the steps, and the `on-*.yml` workflows decide when they run.

Rules for CI:
- **A run's results belong to the commit that started it.** A run may change the code before checking it (apply a formatter, generate, build, bundle) and may push what it produced; those commits are its output, covered by its results
- **Changes made by automation arrive as PRs.** Whatever makes them, an agent, a scheduled workflow or a bot, opens its own PR rather than pushing to another PR's branch, and that PR runs CI like any other. A PR opened with the workflow token starts no workflows, so automation uses a token that does (e.g. a GitHub App's)
- **Nothing skips CI.** Commit messages cannot skip it (§3.2)
- **Tool versions are pinned in the Makefile, and CI runs `make`**, so local runs and CI judge the same way

GitHub's required checks are a fixed list, while which checks apply depends on the change: a required check whose workflow did not start for its paths waits forever. So only the commit message check, which starts for every PR, is required. A human merges until shampoo does; shampoo then decides from the changed paths which checks the head must have passed (design doc §10).

Dependabot, when added, uses the commit prefix `:arrow_up:` so that its commits follow the commit rules, with `bump` as an allowed verb. The exact message it produces is to be verified then.

## 8. Open problems
Recorded here so that they are solved before shampoo runs HOOTL:
- **The upper bound of a minimal commit** (§3.1) is kept by judgment, which needs a human or an agent's discretion
