Feature: Run a task
  `shampoo run` takes a task from the project's task source, creates a
  worktree on its own branch, launches an agent there with the task's
  definition of done, and when the agent ends, judges task done. It exits
  as `shampoo guard task` does.

  Background:
    Given a git repository with a commit

  @skip
  Scenario: An agent works on a task in its own worktree
    Given a task "add-readme" with the definition of done "Add a README"
    And an agent that records where it works
    And the changes are committed
    When I run "shampoo run"
    Then the agent received "Add a README"
    And the agent worked on the branch "task/add-readme"
    And the agent worked outside the repository's own working tree

  @skip
  Scenario: A task whose guard passes is done
    Given a task "add-readme" with the definition of done "Add a README"
    And an agent that changes "README.md"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And the changes are committed
    When I run "shampoo run"
    Then the exit code is 0

  @skip
  Scenario: A task whose guard finds a violation is sent back
    Given a task "add-readme" with the definition of done "Add a README"
    And an agent that changes "README.md"
    And a check "lint" that exits 1
    And a rule guarding task with check "lint"
    And the changes are committed
    When I run "shampoo run"
    Then the exit code is 1
