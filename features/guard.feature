Feature: Guard a transition
  `shampoo guard <transition>` selects the rules that guard the transition
  and whose paths cover what changed, runs each of their checks once in the
  order of their needs, and judges the rules after each check. It stops
  when a judgment reaches its rule's stop. It exits 0 on pass, 1 on a
  task failure, 2 on a workflow failure, 3 when a stop sends the work to
  harness repair, and 4 when failed repairs exceed their limit. With
  `--json` it exits 0 and answers an agent's hook in one JSON object
  carrying every agent's fields.

  Background:
    Given a git repository with a commit

  @sha:fefade2e2479
  Scenario: A passing check lets the edit through
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:0cbc8fed24a4
  Scenario: A violation sends the edit back
    Given a check "lint" that prints "found TODO" and exits 1
    And a rule "no-todo" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo"
    And stderr contains "found TODO"

  @sha:7e7b2cd6655c
  Scenario: A misbehaving check is a workflow failure
    Given a check "lint" that exits 2
    And a rule guarding edit with check "lint" judging exit 2 as a workflow failure
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 2

  @sha:9043b0b89519
  Scenario: A judge command deciding a violation sends the edit back
    Given a check "test" that exits 0
    And a rule "coverage" guarding edit with check "test" whose judge prints "too low" and exits 1
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

  @sha:62697e11cb9a
  Scenario: A rule whose paths miss the change is not checked
    Given a check "lint" that exits 1
    And a rule guarding edit for paths "**.go" with check "lint"
    And a changed file "README.md"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:b7d57b6bd7cf
  Scenario: A rule guarding another transition is not checked
    Given a check "lint" that exits 1
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

  @sha:f7d07e2c75ec
  Scenario: A violation within its budget does not stop the checks after it
    Given a check "build" that exits 1
    And a rule guarding edit with check "build"
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And the check "acceptance" has run once

  @sha:d3c9f18f1743
  Scenario: A violation exceeding a zero budget stops the checks after it
    Given a check "build" that exits 1
    And a rule guarding edit with check "build" and a zero budget
    And a check "acceptance" that needs "build" and records that it ran
    And a rule guarding edit with check "acceptance"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And the check "acceptance" has not run

  @sha:0f83e4f59d23
  Scenario: Every rule on a check is judged before stopping
    Given a check "lint" that exits 1
    And a rule "strict" guarding edit with check "lint" and a zero budget
    And a rule "lenient" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "strict"
    And stderr contains "lenient"

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

  @sha:e68a4f5d750f
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

  @sha:5e1af9e95f01
  Scenario: A budget burnt through the window is reported
    Given a check "lint" that exits 1
    And a rule "no-todo" guarding edit with check "lint" and a budget of 1 violation in 2 runs
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1
    And stderr contains "no-todo burns its budget"

  @sha:90fa3c0a6750
  Scenario: Failed repairs beyond their limit hand the task to a human
    Given settings allowing 1 failed attempt in harness repair
    And a check "lint" that exits 1
    And a rule guarding edit with check "lint" and a zero budget
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 4

  @sha:98e6ba68126e
  Scenario: A merged task is summarized into statistics
    Given a branch "task-a"
    And a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    And "shampoo guard edit" has passed
    And the branch "task-a" is committed and merged
    When I run "shampoo guard edit"
    Then the statistics have a line for task "task-a"

  @sha:c4a489b31d31
  Scenario: An unchanged tree after a failure is not checked again
    Given a check "lint" that records that it ran and exits 1
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    And "shampoo guard edit" has run
    When I run "shampoo guard edit"
    Then the exit code is 1
    And the check "lint" has run once

  @sha:1058bc79c62e
  Scenario: A rule guarding task is checked when the task is done
    Given a branch "task-a"
    And a check "lint" that exits 1
    And a rule guarding task with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard task"
    Then the exit code is 1

  @sha:16db287a50df
  Scenario: A task's committed changes are checked when the task is done
    Given a branch "task-a"
    And a check "lint" that records that it ran
    And a rule guarding task with check "lint"
    And a changed file "main.go"
    And the changes are committed
    When I run "shampoo guard task"
    Then the check "lint" has run once

  @sha:065e5b2a73fd
  Scenario: Violations at edit do not count toward the stop at task
    Given a check "lint" that exits 1
    And a rule "no-todo" guarding edit and task with check "lint" and a stop at task of 2 violations in the latest 2 runs
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    And "shampoo guard edit" has run
    When I run "shampoo guard task"
    Then the exit code is 1

  @sha:62568703dfe5
  Scenario: The default stop at edit lets a change fail on its way
    Given a check "lint" that exits 1
    And a rule guarding edit with check "lint" and the default budgets
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit"
    Then the exit code is 1

  @skip
  Scenario: A violation answers a hook in JSON that sends the agent back
    Given a check "lint" that prints "found TODO" and exits 1
    And a rule "no-todo" guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit --json"
    Then the exit code is 0
    And the answer has "code" 1
    And the answer has "decision" "block"
    And the answer's "reason" contains "found TODO"
    And the answer's "followup_message" contains "found TODO"
    And the answer's "additional_context" contains "found TODO"

  @skip
  Scenario: A pass answers a hook in JSON that lets the agent go on
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint"
    And a changed file "main.go"
    When I run "shampoo guard edit --json"
    Then the exit code is 0
    And the answer has "code" 0
    And the answer has no "decision"
    And the answer has no "followup_message"

  @skip
  Scenario: A workflow failure answers a hook in JSON that shows it to the human
    Given a check "lint" that exits 2
    And a rule guarding edit with check "lint" judging exit 2 as a workflow failure
    And a changed file "main.go"
    When I run "shampoo guard edit --json"
    Then the exit code is 0
    And the answer has "code" 2
    And the answer has no "decision"
    And the answer's "systemMessage" contains "workflow-failure"

  @skip
  Scenario: Repairs beyond their limit answer a hook in JSON that shows it to the human
    Given settings allowing 1 failed attempt in harness repair
    And a check "lint" that exits 1
    And a rule guarding edit with check "lint" and a zero budget
    And a changed file "main.go"
    And "shampoo guard edit" has run
    And a changed file "main.go"
    When I run "shampoo guard edit --json"
    Then the exit code is 0
    And the answer has "code" 4
    And the answer has no "decision"
    And the answer's "systemMessage" contains "hand the task to a human"
