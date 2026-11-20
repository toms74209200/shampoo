Feature: Charge a check's run time to its time budget
  Each run of a check is timed against the limit of its test size. A run
  longer than the limit is a charge to the check's time budget, counted
  over the check's latest runs. When the charges reach the stop, the guard
  sends the work to harness repair, to reorganize the check.

  Background:
    Given a git repository with a commit
    And a check "slow" that takes 1.5 seconds and exits 0
    And a rule guarding edit with check "slow"

  @skip
  @sha:0490f4d89b0d
  Scenario: A run within its size's limit is not charged
    Given a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  @sha:d8bd6f916a35
  Scenario: One run over the limit lets the work through
    Given settings limiting small checks to 1 second
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  @sha:2aee7970dea3
  Scenario: Runs over the limit reaching the stop send the work to harness repair
    Given settings limiting small checks to 1 second
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "slow exceeds its time budget"
    And stderr contains "reorganize the check"

  @skip
  @sha:427f294faf04
  Scenario: A limit that is not positive is a workflow failure
    Given settings limiting small checks to 0 seconds
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2
