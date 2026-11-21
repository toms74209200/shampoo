Feature: Review a task against its policies
  When a policy is marked for review, a task whose guard has passed is
  reviewed before its pull request. `shampoo next` says to review, with
  the marked policies and the project's requirements on the review. The
  reviewing agent records its findings with `shampoo review`. Findings
  send the task back to its editing, with them, and it is reviewed again
  once changed; a review without findings lets it open its pull request.
  Findings that do not drop from one review to the next show a policy
  that does not settle what to do, and send the work to repair it.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"
    And a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And a changed file "main.go"

  @sha:47b32f4db431
  Scenario: Without a policy for review, a passed task opens a pull request
    Given a policy "layout"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"

  @sha:a9a17e9269f6
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

  @sha:84b6830dacb6
  Scenario: The requirements on the review are handed over
    Given a policy "functions" for review
    And settings requiring "model" "claude-opus-5-5" of the review
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "review"
    And the next action has the requirement "model" "claude-opus-5-5"

  @sha:aeef627ba636
  Scenario: A review with findings sends the task back with them
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found "a wrapper used once" on policy "functions"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "continue"
    And the next action has "branch" "task-a"
    And the next action has the finding "a wrapper used once" on policy "functions"

  @sha:1fc9463d7c9e
  Scenario: A review without findings lets the task open a pull request
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found nothing
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"
    And the next action has no findings

  @sha:e2cf89c81ca9
  Scenario: A task changed after a review with findings is reviewed again
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found "a wrapper used once" on policy "functions"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "review"

  @sha:c0ed05b8634f
  Scenario: A task changed after a review without findings is not reviewed again
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found nothing
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "open-pr"

  @sha:557091e79251
  Scenario: Findings that drop from one review to the next keep the task going
    Given a policy "functions" for review
    And settings allowing 0 reviews whose findings do not drop
    And "shampoo guard task" has passed
    And the review found 2 findings on policy "functions"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And the review found 1 finding on policy "functions"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "continue"

  @sha:03d06d8c939d
  Scenario: Findings that do not drop repair the harness
    Given a policy "functions" for review
    And settings allowing 0 reviews whose findings do not drop
    And "shampoo guard task" has passed
    And the review found 1 finding on policy "functions"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And the review found 1 finding on policy "functions"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "repair-harness"
    And the next action has "branch" "task-a"
    And the next action has the policies "functions"

  @sha:154a9dc4cf1c
  Scenario: A review of another task does not count
    Given a policy "functions" for review
    And "shampoo guard task" has passed
    And the review found nothing
    And the changes are committed
    And a branch "task-b"
    And a changed file "lib.go"
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "review"
    And the next action has "branch" "task-b"

  @sha:f10a35906ceb
  Scenario: A finding on a policy not marked for review is a workflow failure
    Given a policy "layout"
    And a policy "functions" for review
    And "shampoo guard task" has passed
    When I run "shampoo review" with a finding on policy "layout"
    Then the exit code is 2

  @sha:ee438d8f661f
  Scenario: Findings that do not drop count against the repair budget
    Given a policy "functions" for review
    And settings allowing 0 reviews whose findings do not drop
    And settings allowing 0 failed attempt in harness repair
    And "shampoo guard task" has passed
    And the review found 1 finding on policy "functions"
    And a changed file "main.go"
    And "shampoo guard task" has passed
    And the review found 1 finding on policy "functions"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "hand-to-human"

  @skip
  Scenario: A task whose commits changed after a review with findings is reviewed again
    Given a policy "pull-requests" for review
    And the changes are committed
    And "shampoo guard task" has passed
    And the review found "one commit mixes two changes" on policy "pull-requests"
    And the commits are made again with the same tree
    And "shampoo guard task" has passed
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "review"
