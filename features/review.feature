Feature: Review a task against its policies
  When a policy is marked for review, a task whose guard has passed is
  reviewed once before its pull request. `shampoo next` says to review,
  with the marked policies and the project's requirements on the review.
  The reviewing agent records its findings with `shampoo review`, and
  `shampoo next` hands them to the agent that did the work, with the
  action to open the pull request, as feedback.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And a changed file "main.go"

  @skip
  Scenario: Without a policy for review, a passed task opens a pull request
    Given a policy "layout"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"

  @skip
  Scenario: A passed task with a policy for review is reviewed first
    Given a policy "layout"
    And a policy "functions" for review
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "review"
    And the next action has "branch" "task-a"
    And the next action has the policies "functions"

  @skip
  Scenario: The requirements on the review are handed over
    Given a policy "functions" for review
    And settings requiring "model" "claude-opus-5-5" of the review
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "review"
    And the next action has the requirement "model" "claude-opus-5-5"

  @skip
  Scenario: A reviewed task opens a pull request with the findings
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found "a wrapper used once" on policy "functions"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"
    And the next action has the finding "a wrapper used once" on policy "functions"

  @skip
  Scenario: A review without findings lets the task open a pull request
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found nothing
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"
    And the next action has no findings

  @skip
  Scenario: A task changed after its review is not reviewed again
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found "a wrapper used once" on policy "functions"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"
    And the next action has no findings

  @skip
  Scenario: A finding on a policy not marked for review is a workflow failure
    Given a policy "layout"
    And a policy "functions" for review
    And "shampoo guard task" has passed
    When I run "shampoo review" with a finding on policy "layout"
    Then the exit code is 2
