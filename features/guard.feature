Feature: Guard a transition
  `shampoo guard <transition>` selects the rules that guard the transition
  and whose paths cover what changed, runs each of their checks once, after
  the checks it needs, and judges each rule. It exits 0 on pass, 1 on a
  task failure and 2 on a workflow failure.

  Background:
    Given a git repository with a commit

  @sha:fefade2e2479
  Scenario: A passing check lets the edit through
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:d87bb6163dad
  Scenario: A failing check sends the edit back
    Given a check "lint" that prints "found TODO" and exits 1 as a task failure
    And a rule "no-todo" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo"
    And stderr contains "found TODO"

  @sha:328c0e0fc1b7
  Scenario: A misbehaving check is a workflow failure
    Given a check "lint" that exits 2 as a workflow failure
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2

  @sha:98e6415a5e36
  Scenario: A failing judge sends the edit back
    Given a check "test" that exits 0
    And a rule "coverage" guarding edit with check "test" whose judge prints "too low" and exits 1 as a task failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "coverage"
    And stderr contains "too low"

  @sha:bd5660afc27d
  Scenario: A judge charging its budget lets the edit through
    Given a check "test" that exits 0
    And a rule guarding edit with check "test" whose judge exits 3 charging its budget
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:90c4bf029195
  Scenario: A rule whose paths miss the change is not checked
    Given a check "lint" that exits 1 as a task failure
    And a rule guarding edit for paths "**.go" with check "lint"
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:4a30c6c9bab7
  Scenario: A rule guarding another transition is not checked
    Given a check "lint" that exits 1 as a task failure
    And a rule guarding pr with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:5491bdd775c8
  Scenario: Rules pointing to the same check share one run
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the check "test" has run once

  @sha:b417c4a066bf
  Scenario: A check runs after the checks it needs
    Given a check "build" that records that it ran
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the checks ran in the order "build", "acceptance"

  @sha:3f8b26a05598
  Scenario: A check whose needed check failed is not run
    Given a check "build" that exits 1 as a task failure
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And the check "acceptance" has not run

  @sha:8a0d183c5d4b
  Scenario: Nothing changed since the last pass runs no check
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    When I run "shampoo guard edit"
    Then the check "test" has run once

  @sha:9a63ec93bf2f
  Scenario: A change after the last pass is checked
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the check "test" has run twice
