Feature: Count the workflow failures of a rule
  A workflow failure is a judgment the check or its judge could not make.
  It does not send the work back. The guard counts a rule's workflow
  failures over its latest runs, apart from its violations, and when they
  reach the stop, sends the work to harness repair.

  Background:
    Given a git repository with a commit
    And a check "lint" that exits 2
    And a rule guarding edit with check "lint" judging exit 2 as a workflow failure

  @sha:9a62f0cefb1d
  Scenario: One workflow failure does not stop the work
    Given a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2

  @sha:1ca99a9df741
  Scenario: Workflow failures reaching the stop send the work to harness repair
    Given a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "exceeds its workflow-failure budget"
    And stderr contains "repair the harness"

  @sha:8522e186ccae
  Scenario: A project's stop overrides the default
    Given settings stopping workflow failures at 1 in the latest 1 run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3

  @sha:3a792ad37936
  Scenario: A stop with more workflow failures than runs is a workflow failure of the settings
    Given settings stopping workflow failures at 2 in the latest 1 run
    And a check "build" that exits 0
    And a rule guarding edit with check "build"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2
    And stderr contains "invalid settings"
