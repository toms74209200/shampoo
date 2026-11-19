Feature: Guard a transition
  `shampoo guard <transition>` selects the rules that guard the transition
  and whose paths cover what changed, runs each of their checks once in the
  order of their needs, and judges the rules after each check. It stops
  when a judgment reaches its rule's stop. It exits 0 on pass, 1 on a
  task failure, 2 on a workflow failure, 3 when a stop sends the work to
  harness repair, and 4 when failed repairs exceed their limit.

  Background:
    Given a git repository with a commit

  @skip
  Scenario: A passing check lets the edit through
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A violation sends the edit back
    Given a check "lint" that prints "found TODO" and exits 1
    And a rule "no-todo" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo"
    And stderr contains "found TODO"

  @skip
  Scenario: A misbehaving check is a workflow failure
    Given a check "lint" that exits 2
    And a rule guarding edit with check "lint" judging exit 2 as a workflow failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2

  @skip
  Scenario: A judge command deciding a violation sends the edit back
    Given a check "test" that exits 0
    And a rule "coverage" guarding edit with check "test" whose judge prints "too low" and exits 1
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "coverage"
    And stderr contains "too low"

  @skip
  Scenario: A judge charging its budget lets the edit through
    Given a check "test" that exits 0
    And a rule guarding edit with check "test" whose judge exits 3 charging its budget
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A rule whose paths miss the change is not checked
    Given a check "lint" that exits 1
    And a rule guarding edit for paths "**.go" with check "lint"
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: A rule guarding another transition is not checked
    Given a check "lint" that exits 1
    And a rule guarding pr with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @skip
  Scenario: Rules pointing to the same check share one run
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the check "test" has run once

  @skip
  Scenario: A check runs after the checks it needs
    Given a check "build" that records that it ran
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the checks ran in the order "build", "acceptance"

  @skip
  Scenario: A violation within its budget does not stop the checks after it
    Given a check "build" that exits 1
    And a rule guarding edit with check "build"
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And the check "acceptance" has run once

  @skip
  Scenario: A violation exceeding a zero budget stops the checks after it
    Given a check "build" that exits 1
    And a rule guarding edit with check "build" and a zero budget
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And the check "acceptance" has not run

  @skip
  Scenario: Every rule on a check is judged before stopping
    Given a check "lint" that exits 1
    And a rule "strict" guarding edit with check "lint" and a zero budget
    And a rule "lenient" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "strict"
    And stderr contains "lenient"

  @skip
  Scenario: Nothing changed since the last pass runs no check
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    When I run "shampoo guard edit"
    Then the check "test" has run once

  @skip
  Scenario: A change after the last pass is checked
    Given a check "test" that records that it ran
    And a rule guarding edit with check "test"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the check "test" has run twice

  @skip
  Scenario: Violations burning a budget fast send the work to harness repair
    Given a check "lint" that exits 1
    And a rule "no-todo" guarding edit with check "lint" and a stop at 2 violations in the latest 2 runs
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "no-todo exceeds its budget"
    And stderr contains "repair the harness"

  @skip
  Scenario: A budget burnt through the window is reported
    Given a check "lint" that exits 1
    And a rule "no-todo" guarding edit with check "lint" and a budget of 1 violation in 2 runs
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo burns its budget"

  @skip
  Scenario: Failed repairs beyond their limit hand the task to a human
    Given settings allowing 1 failed attempt in harness repair
    And a check "lint" that exits 1
    And a rule guarding edit with check "lint" and a zero budget
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 4

  @skip
  Scenario: A merged task is summarized into statistics
    Given a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics have a line for task "task-a"
