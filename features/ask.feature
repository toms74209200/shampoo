Feature: Ask a human
  An agent that cannot settle a decision, or that the policies tell to,
  decides, goes on, and asks a human with `shampoo ask`: what it decided,
  how else it could be built, and the trade-offs. shampoo opens a GitHub
  issue and records it as a question of the task, without waiting for the
  answer. It needs no label: it tells its issues by its record.

  Background:
    Given a git repository with a commit
    And a branch "task-a"

  @skip
  @sha:471e34250a02
  Scenario: A question opens an issue and records it
    When I run "shampoo ask" deciding "Merge by squash" instead of "a merge commit" for "a linear history"
    Then the exit code is 0
    And stdout contains "https://github.example/issues/1"
    And an issue without a label is open titled "Merge by squash"
    And the issue is recorded as a question of the task "task-a"
    And the issue says "a merge commit"
    And the issue says "a linear history"
    And the issue says "task-a"

  @sha:4ea4a408d347
  Scenario: A question without alternatives is a workflow failure
    When I run "shampoo ask" deciding "Merge by squash" instead of "" for "a linear history"
    Then the exit code is 2
    And no issue is open

  @sha:6a7a2474dfe5
  Scenario: A question is counted for its task
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And "shampoo ask" has asked "Merge by squash"
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics have 1 question for task "task-a"
