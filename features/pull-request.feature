Feature: Follow a pull request to its merge
  Once a task's guard has passed and its pull request is open,
  `shampoo next` follows the pull request's checks. The expected checks
  are the workflows whose `pull_request` filters match the pull request's
  changed paths. For each, the latest run on any of the pull request's
  commits counts. A missing or running check means wait, a failed one
  means fix, and when every expected check has passed, shampoo merges the
  pull request itself.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And a changed file "main.go"

  @skip
  Scenario: A pull request whose expected workflow is running waits
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" is running on the latest commit
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "wait"
    And the next action has "branch" "task-a"

  @skip
  Scenario: A pull request with no run of an expected workflow waits
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    When I run "shampoo next"
    Then the next action is "wait"

  @skip
  Scenario: A workflow without paths is expected for every change
    Given a workflow "go" on pull requests to paths "**.go"
    And a workflow "commit" on pull requests
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has succeeded on the latest commit
    When I run "shampoo next"
    Then the next action is "wait"

  @skip
  Scenario: A workflow whose paths the change does not touch is not expected
    Given a workflow "go" on pull requests to paths "**.go"
    And a workflow "docs" on pull requests to paths "docs/**"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has succeeded on the latest commit
    When I run "shampoo next"
    Then the next action is "merged"

  @skip
  Scenario: A pull request whose expected workflow failed is fixed
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has failed on the latest commit
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "fix-pr"
    And the next action has "branch" "task-a"
    And the next action names the failed workflow "go" with its run's URL
    And the pull request is not merged

  @skip
  Scenario: A pull request whose expected workflows passed is merged
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has succeeded on the latest commit
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "merged"
    And the next action has "branch" "task-a"
    And the pull request is merged with "--merge"

  @skip
  Scenario: A run on an earlier commit of the pull request counts
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has succeeded on the latest commit
    And the run pushed a commit to the pull request
    When I run "shampoo next"
    Then the next action is "merged"

  @skip
  Scenario: The latest run of a workflow counts
    Given a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has failed on the latest commit
    And the workflow "go" has succeeded on the latest commit
    When I run "shampoo next"
    Then the next action is "merged"

  @skip
  Scenario: The merge method follows the settings
    Given settings merging by "squash"
    And a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" has succeeded on the latest commit
    When I run "shampoo next"
    Then the pull request is merged with "--squash"

  @skip
  Scenario: Waiting longer than the limit is a workflow failure
    Given settings waiting at most 0 minutes for checks
    And a workflow "go" on pull requests to paths "**.go"
    And "shampoo guard task" has passed
    And the changes are committed
    And a pull request is open for the branch
    And the workflow "go" is running on the latest commit
    When I run "shampoo next"
    Then the exit code is 2
    And the pull request is not merged
