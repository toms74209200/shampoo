---
name: shampoo-start
description: Work on a repository set up for shampoo, through `shampoo next` (docs/design.md §11). Use at the start of a session in such a repository, or when asked to go on with the work, take the next task, or run the loop.
---

# Working through shampoo

`shampoo next` says what to do next, as JSON with an `action`. Ask it, do what
it says, and ask again. Do not keep state of your own: it looks at git, GitHub
and the ledger each time.

| Action | What to do |
|---|---|
| `take-task` | Read the task source in `source`, take one task, make a branch for it from the default branch, and work on it |
| `continue` | Go on with the task on `branch`; the guard sends you back with the reasons |
| `repair-harness` | The same violations keep coming: find whether the work or the check causes them, and fix that before the task |
| `review` | Review the change against the `policies`, under the `requirements`, and record it with `shampoo review` |
| `open-pr` | Push the branch and open the pull request; put the `findings` in its description |
| `wait` | Wait for the checks of the pull request, then ask again; do not poll faster than they run |
| `fix-pr` | Read each failed run in `failures`, fix the cause, and push |
| `merged` | The task is done: mark it done in the task source, go back to the default branch, and ask again |
| `review-workflow` | Look at the policies, the settings and the checks for what made the tasks in `asked` and `repaired` ask a human or repair the harness, and change what needs changing, as a task |
| `wait-for-human` | Stop: questions in `questions` wait for a human |
| `hand-to-human` | Stop: the task on `branch` is handed to a human |
| `set-task-source` | Stop: the task source is not set or not found |

When you cannot settle a decision, decide, go on, and ask with `shampoo ask`,
as the policies in your context say.
