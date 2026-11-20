Feature: Say what to do next
  `shampoo next` looks at the current state and prints the next action as
  a JSON object, such as `{"action": "open-pr", "branch": "task-a"}`, for
  the agents and for the scripts running them.

  Background:
    Given a git repository with a commit

  @skip
  Scenario: With no task in progress, a task is taken from the source
    Given a task source "todo.md"
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "take-task"
    And the next action has "source" "todo.md"

  @skip
  Scenario: A task source that does not exist is a workflow failure
    Given a task source "todo.md" that does not exist
    When I run "shampoo next"
    Then the exit code is 2

  @skip
  Scenario: A task whose guard has not passed goes on
    Given a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 1
    And a rule guarding task with check "lint"
    And a changed file "main.go"
    And "shampoo guard task" has run
    When I run "shampoo next"
    Then the next action is "continue"
    And the next action has "branch" "task-a"

  @skip
  Scenario: A task whose guard passed opens a pull request
    Given a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"
    And the next action has "branch" "task-a"

  @skip
  Scenario: A task over its rule's budget repairs the harness
    Given a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 1
    And a rule guarding task with check "lint" and a zero budget
    And a changed file "main.go"
    And "shampoo guard task" has run
    When I run "shampoo next"
    Then the next action is "repair-harness"
    And the next action has "branch" "task-a"

  @skip
  Scenario: A task over its limit of failed repairs is handed to a human
    Given a task source "todo.md"
    And settings allowing 0 failed attempt in harness repair
    And a branch "task-a"
    And a check "lint" that exits 1
    And a rule guarding task with check "lint" and a zero budget
    And a changed file "main.go"
    And "shampoo guard task" has run
    When I run "shampoo next"
    Then the next action is "hand-to-human"
    And the next action has "branch" "task-a"
