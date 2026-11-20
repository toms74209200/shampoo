Feature: Ask a human
  An agent that cannot settle a decision, or that the policies tell to,
  decides, goes on, and asks a human with `shampoo ask`: what it decided,
  how else it could be built, and the trade-offs. shampoo opens a GitHub
  issue with the label for questions and records the question for the
  task, without waiting for the answer.

  Background:
    Given a git repository with a commit
    And a branch "task-a"

  @skip
  @sha:e45133b166eb
  Scenario: A question opens a labeled issue
    When I run "shampoo ask" deciding "Merge by squash" instead of "a merge commit" for "a linear history"
    Then the exit code is 0
    And stdout contains "https://github.example/issues/1"
    And an issue labeled "human-review" is open titled "Merge by squash"
    And the issue says "a merge commit"
    And the issue says "a linear history"
    And the issue says "task-a"

  @skip
  @sha:7be9ad8b2526
  Scenario: The label comes from the settings
    Given settings asking a human with the label "ask" and at most 3 open questions
    When I run "shampoo ask" deciding "Merge by squash" instead of "a merge commit" for "a linear history"
    Then an issue labeled "ask" is open titled "Merge by squash"

  @skip
  @sha:4ea4a408d347
  Scenario: A question without alternatives is a workflow failure
    When I run "shampoo ask" deciding "Merge by squash" instead of "" for "a linear history"
    Then the exit code is 2
    And no issue is open

  @skip
  @sha:6a7a2474dfe5
  Scenario: A question is counted for its task
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And "shampoo ask" has asked "Merge by squash"
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics have 1 question for task "task-a"
