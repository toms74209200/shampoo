Feature: Count the exceptions of a check
  A rule that allows exceptions declares the file its check writes them
  to, one line per exception, and their limit. After the check runs, the
  guard counts the lines. More exceptions than the limit send the work to
  harness repair, to fix the check rather than add exceptions.

  Background:
    Given a git repository with a commit
    And a changed file "main.go"

  @sha:1afb465104b5
  Scenario: Exceptions within the default limit let the work through
    Given a check "lint" that writes 3 exceptions to "out/exceptions.txt"
    And a rule guarding edit with check "lint" and its exceptions in "out/exceptions.txt"
    When I run "shampoo guard edit"
    Then the exit code is 0

  @sha:6b39c42a36c0
  Scenario: Exceptions over the default limit send the work to harness repair
    Given a check "lint" that writes 4 exceptions to "out/exceptions.txt"
    And a rule guarding edit with check "lint" and its exceptions in "out/exceptions.txt"
    When I run "shampoo guard edit"
    Then the exit code is 3
    And stderr contains "4 exceptions"
    And stderr contains "fix the check"

  @sha:aebd0f14c824
  Scenario: A rule's limit overrides the default
    Given a check "lint" that writes 2 exceptions to "out/exceptions.txt"
    And a rule guarding edit with check "lint" and its exceptions in "out/exceptions.txt" limited to 1
    When I run "shampoo guard edit"
    Then the exit code is 3

  @sha:ec5e1953f652
  Scenario: A check that does not write its exceptions is a workflow failure
    Given a check "lint" that exits 0
    And a rule guarding edit with check "lint" and its exceptions in "out/exceptions.txt"
    When I run "shampoo guard edit"
    Then the exit code is 2
