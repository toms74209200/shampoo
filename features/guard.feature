Feature: Guard a transition
  `shampoo guard <transition>` runs the checks of the rules that guard the
  transition and whose paths cover what changed, and exits 0 on pass, 1 on
  a task failure and 2 on a workflow failure.

  Background:
    Given a git repository with a commit

  @skip
  Scenario: A passing check lets the edit through
    Given a rule guarding edit whose check exits 0
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A failing check sends the edit back
    Given a rule "no-todo" guarding edit whose check prints "found TODO" and exits 1 as a task failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo"
    And stderr contains "found TODO"

  @skip
  Scenario: A misbehaving check is a workflow failure
    Given a rule guarding edit whose check exits 2 as a workflow failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2

  @skip
  Scenario: A check charging its budget lets the edit through
    Given a rule guarding edit whose check exits 3 charging its budget
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A rule whose paths miss the change is not checked
    Given a rule guarding edit for paths "**.go" whose check exits 1 as a task failure
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A rule guarding another transition is not checked
    Given a rule guarding pr whose check exits 1 as a task failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: Nothing changed since the last pass runs no check
    Given a rule guarding edit whose check records that it ran
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    When I run "shampoo guard edit"
    Then the check has run once

  @skip
  Scenario: A change after the last pass is checked
    Given a rule guarding edit whose check records that it ran
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the check has run twice
