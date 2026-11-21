---
name: shampoo-init
description: Set up shampoo in a repository, up to the point where it runs (docs/design.md §7.1). Use when asked to adopt, install or set up shampoo in a project, from a brownfield repository or from a reference repository.
---

# Setting up shampoo

You set up shampoo in the target repository. Decide what the repositories
settle, ask only what they do not, and leave the project with a guard that
passes.

## 1. Find out what the project already does

Look at the target repository first, then at the reference repository if one
is given. A repository shows its checks in what it does, whether it uses
shampoo or not:

- Its linter and the linter's configuration, and how it is run (a Makefile, package scripts, a pinned script)
- How it runs its tests, and which are small (no network, no files outside a temporary directory, no processes)
- How it measures coverage, and any threshold it already keeps
- Its CI workflows in `.github/workflows/`, and how pull requests merge
- Its commit conventions, from `git log` and any commit message check

When the reference repository has `.shampoo/`, carry its rules, checks and
policies over first, adapted to the target, including rules that are not
recommended. `shampoo init` keeps every file that exists, so what you carried
over stays.

## 2. Decide, and ask only what is left open

Decide everything the repositories settle. Ask the human only where they leave
a choice open, or where the change reaches beyond the project's own files,
such as changing its CI or the way pull requests merge. Ask all such questions
at once, each with your proposal, so that setting up is not a dialogue.

When told to set everything up hands-off, ask nothing: decide it all, and
write what you decided in the description of the pull request that adds the
setup.

## 3. Assemble the checks of the recommended rules

The recommended rules say what to check, not how. Give a command for each one
the project can have:

| Rule | The command |
|---|---|
| `lint` | Runs the project's linter; exits 1 on a violation. If the linter lets violations pass by an exception in the code, write each exception as one line to a file and give that file as `exceptions` |
| `small-tests` | Runs the small tests; exits 1 when one fails |
| `coverage` | Measures small-test coverage; exits 1 below the threshold (80% is a start) |
| `commit-message` | Checks the message of the commit given as `<commit>`; exits 1 when it does not follow the format |

A command exits 1 for a problem in the work and any other non-zero code for a
problem in the check itself. Write a script under `scripts/` when the command
needs more than one line. Give `paths` when a check only concerns some files,
such as `["**.go", "go.mod"]`.

## 4. Run `shampoo init`

    echo '{"tasks": "<task source>", "checks": {"lint": {"command": "...", "exceptions": "..."}, "small-tests": {"command": "..."}, "coverage": {"command": "..."}, "commit-message": {"command": "... <commit>"}}}' | shampoo init

It writes the rules, checks, recommended policies, the task source and the
hooks in `.claude/settings.json`. It needs no label on GitHub.
It prints what it wrote and what it kept. When it says there is no review
phase, no policy is marked for review: keep it so only if the project wants no
review.

## 5. See it run

Run `shampoo guard task` and fix what fails: a failing check on code that is
fine is a check to fix, not code. Then commit the setup on a branch of its own
and open a pull request with what you decided.
