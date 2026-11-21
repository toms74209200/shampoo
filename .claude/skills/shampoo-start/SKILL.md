---
name: shampoo-start
description: Work on a repository set up for shampoo through `shampoo next`. Use at session start in a shampoo repository, or when asked to continue work, take the next task, or run the loop.
---

# Working through shampoo

Run `shampoo next`, do the `action` it returns as JSON, and run `shampoo next`
again. Keep no state of your own. If `shampoo next` exits non-zero, treat it as
a workflow failure to fix or report rather than guessing.

| Action | What to do |
|---|---|
| `take-task` | Pick one task from `source`, create a branch from the default branch (`git switch -c`), and work until its "Done when" holds. In `tasks/`, delete the task file in the same PR. When no tasks remain, stop |
| `continue` | Continue on `branch`; hooks run `shampoo guard` and send violations back. With `findings`, address them — once the tree changes and passes `shampoo guard task`, `shampoo next` will ask for another review |
| `repair-harness` | Without `policies`, find whether the work or the check caused the repeated violations, and fix that before going on with the task. With `policies`, review findings did not converge: clarify those `.shampoo/policies/` files so the work and the review agree |
| `review` | Review the changes on `branch` against `policies` (under `requirements` if set), and record the result with `shampoo review` |
| `open-pr` | Push `branch` and open a PR (`gh pr create`), then run `shampoo next` |
| `wait` | Wait for the PR's checks without polling faster than CI runs, then run `shampoo next` |
| `fix-pr` | Inspect each failed run in `failures` (`workflow`, `url`), fix the cause on `branch`, and push |
| `merged` | The PR was merged: switch back to the default branch, pull, and run `shampoo next` |
| `review-workflow` | Inspect `.shampoo/policies/`, `.shampoo/settings.toml`, and `.shampoo/checks/` for why the tasks in `asked` or `repaired` asked a human or repaired the harness too often, and fix the cause as a task PR |
| `wait-for-human` | Stop: the GitHub issues in `questions` await a human |
| `hand-to-human` | Stop: the task on `branch` is handed to a human |

Record a review (`[]` when nothing is found):

```sh
echo '{"findings": [{"policy": "<id>", "comment": "..."}]}' | shampoo review
```

Ask a human without waiting when a decision cannot be settled:

```sh
echo '{"decision": "...", "alternatives": "...", "tradeoffs": "..."}' | shampoo ask
```
