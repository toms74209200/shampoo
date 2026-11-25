Feature: Token usage of a task
  At a stop, the task guard of the hook reads the agent's transcript,
  named by the hook's input, and those of its subagents, and adds the tokens used since
  the previous stop to the task, each kind apart. The statistics of a merged task carry its tokens
  and the lines it changed, the evidence for whether a task cost more than
  its size explains.

  Background:
    Given a git repository with a commit
    And a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding task with check "lint"
    And a changed file "main.go"

  @skip
  @sha:94975e034950
  Scenario: A merged task has the tokens of its transcript by kind
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have 1 input, 2 cache creation, 3 cache read and 4 output tokens

  @skip
  @sha:ea1cd3b7d846
  Scenario: A message written more than once counts once
    Given a transcript with a message written twice using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have 1 input, 2 cache creation, 3 cache read and 4 output tokens

  @skip
  @sha:5f79f13e95e7
  Scenario: A stop counts only what the transcript added since the previous stop
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the transcript adds a message using 10 input, 20 cache creation, 30 cache read and 40 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have 11 input, 22 cache creation, 33 cache read and 44 output tokens

  @skip
  @sha:ec78fe52572a
  Scenario: The next task in the same session does not count the earlier one
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    And a branch "task-b"
    And a changed file "lib.go"
    And the transcript adds a message using 10 input, 20 cache creation, 30 cache read and 40 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-b" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-b" have 10 input, 20 cache creation, 30 cache read and 40 output tokens

  @skip
  @sha:516723ab8428
  Scenario: A merged task has the lines it changed
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have the lines its merge changed

  @skip
  @sha:b3345e680c94
  Scenario: A stop without a transcript records no tokens
    Given "shampoo hook task" has passed
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have no tokens

  @skip
  @sha:87a0556831b6
  Scenario: Tokens used on the default branch count for no task
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    And the transcript adds a message using 10 input, 20 cache creation, 30 cache read and 40 output tokens
    And "shampoo hook task" has passed with the transcript
    And a branch "task-b"
    And a changed file "lib.go"
    And the transcript adds a message using 100 input, 200 cache creation, 300 cache read and 400 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-b" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-b" have 100 input, 200 cache creation, 300 cache read and 400 output tokens
    And the statistics have no line for the default branch

  @skip
  @sha:8b6aedc4684a
  Scenario: The tokens of a subagent count for the task
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And a subagent transcript with a message using 10 input, 20 cache creation, 30 cache read and 40 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have 11 input, 22 cache creation, 33 cache read and 44 output tokens

  @skip
  @sha:af46ab7a0b1b
  Scenario: A subagent transcript counts only what it added since the previous stop
    Given a transcript with a message using 1 input, 2 cache creation, 3 cache read and 4 output tokens
    And a subagent transcript with a message using 10 input, 20 cache creation, 30 cache read and 40 output tokens
    And "shampoo hook task" has passed with the transcript
    And the subagent transcript adds a message using 100 input, 200 cache creation, 300 cache read and 400 output tokens
    And "shampoo hook task" has passed with the transcript
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics of task "task-a" have 111 input, 222 cache creation, 333 cache read and 444 output tokens

  @skip
  @sha:95574c86fdeb
  Scenario: A guard run by hand does not wait for its input
    When I run "shampoo guard task" with an input left open
    Then the guard finishes within 10 seconds
    And the exit code is 0
