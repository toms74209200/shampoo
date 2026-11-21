Feature: Review the workflow when tasks ask a human or repair too often
  Among the latest merged tasks, too many that asked a human or entered
  harness repair mean the workflow itself needs a review: `shampoo next`
  says to review it instead of handing a new task. Reviews that keep
  failing hand the workflow to a human, through an issue.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"

  @sha:0b0578368b64
  Scenario: Tasks asking a human within the limit let a new task be taken
    Given 10 merged tasks that asked a human
    And 10 merged tasks that did not
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:f647f3791f02
  Scenario: More tasks asking a human than the limit review the workflow
    Given 11 merged tasks that asked a human
    And 9 merged tasks that did not
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "review-workflow"
    And the next action names 11 tasks that asked a human

  @sha:457925501354
  Scenario: More tasks repairing the harness than the limit review the workflow
    Given 11 merged tasks that entered harness repair
    When I run "shampoo next"
    Then the next action is "review-workflow"
    And the next action names 11 tasks that repaired the harness

  @sha:218f7d457a07
  Scenario: Only the latest merged tasks count
    Given 11 merged tasks that asked a human
    And 20 merged tasks that did not
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:ce661fd400f2
  Scenario: Tasks merged before the latest review do not count
    Given 11 merged tasks that asked a human
    And "shampoo next" has said to review the workflow
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:5845450c744b
  Scenario: The settings override the window and the limits
    Given settings reviewing the workflow over 2 tasks with at most 0 questions and 0 repairs
    And 1 merged task that asked a human
    When I run "shampoo next"
    Then the next action is "review-workflow"

  @skip
  @sha:2b9cd8cc2b42
  Scenario: Reviews failing more times than the limit hand the workflow to a human
    Given settings allowing 0 failed workflow reviews
    And "shampoo next" has said to review the workflow
    And 11 merged tasks that asked a human
    When I run "shampoo next"
    Then the next action is "wait-for-human"
    And an issue handing the workflow to a human is open
    And the next action names 1 question

  @skip
  @sha:adece96e9e44
  Scenario: An open issue handing the workflow to a human takes no new task
    Given an issue handing the workflow to a human is open
    When I run "shampoo next"
    Then the next action is "wait-for-human"

  @skip
  @sha:0a41dd259625
  Scenario: A closed issue handing the workflow to a human lets the work go on
    Given settings allowing 0 failed workflow reviews
    And "shampoo next" has said to review the workflow
    And 11 merged tasks that asked a human
    And "shampoo next" has handed the workflow to a human
    And the issue handing the workflow to a human is closed
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:966f5c25e7f1
  Scenario: A task already taken goes on
    Given 11 merged tasks that asked a human
    And a branch "task-a"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "continue"
