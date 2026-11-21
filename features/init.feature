Feature: Write the recommended rules for the checks a project decided
  `shampoo init` reads the check commands an agent decided for the
  project, and writes the recommended rules for them, their checks, the
  recommended policies, the task source and the hooks that run shampoo.
  It decides nothing about the project, overwrites nothing and needs no
  label.

  Background:
    Given a git repository with a commit

  @sha:1461a9028b7b
  Scenario: The rule and the check given are written
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the exit code is 0
    And the file ".shampoo/rules/lint.toml" exists
    And the check "lint" runs "make lint"
    And the settings have the task source "todo.md"

  @sha:fdc64a177183
  Scenario: A recommended rule whose check is not given is not written
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the file ".shampoo/rules/small-tests.toml" does not exist

  @sha:9f9981174faa
  Scenario: A project set up by init passes its guard
    Given a task source file "todo.md"
    When shampoo init runs with the task source "todo.md" and the check "small-tests" running "true"
    And I run "shampoo guard task"
    Then the exit code is 0

  @sha:b1b24d47e282
  Scenario: The recommended policies are written for review
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the policy "functions" is marked for review
    And the policy "pull-requests" is marked for review

  @sha:9cc1ffd976d5
  Scenario: Without a policy for review, init says there is no review phase
    Given a policy "functions"
    And a policy "pull-requests"
    And a policy "tests"
    And a policy "comments"
    And a policy "errors"
    And a policy "scope"
    And a policy "reuse"
    And a policy "assertions"
    And a policy "overfitting"
    And a policy "dead-code"
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the exit code is 0
    And stderr contains "no review phase"

  @sha:01a17f6e90c2
  Scenario: The hooks that run shampoo are added
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the Claude Code settings have a "PostToolUse" hook running "shampoo guard edit"
    And the Claude Code settings have a "Stop" hook running "shampoo guard task"
    And the Claude Code settings have a "SessionStart" hook running "shampoo policies"

  @sha:5b7bcafd4c41
  Scenario: The Claude Code settings that exist are kept
    Given Claude Code settings with the permission "Bash(make:*)"
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the Claude Code settings allow "Bash(make:*)"
    And the Claude Code settings have a "Stop" hook running "shampoo guard task"

  @sha:5f571019ce48
  Scenario: The exceptions of the linter are declared on its rule
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint" writing its exceptions to "out/lint.txt"
    Then the rule "lint" declares its exceptions in "out/lint.txt"

  @sha:e616f963dba1
  Scenario: Files that exist are kept
    Given a rule "lint" written by hand
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the exit code is 0
    And the rule "lint" is the one written by hand
    And init reports ".shampoo/rules/lint.toml" as kept

  @sha:f20f2f61095e
  Scenario: No label is created
    When shampoo init runs with the task source "todo.md" and the check "lint" running "make lint"
    Then the exit code is 0
    And no label is created

  @sha:d8186ec052ab
  Scenario: A check for a rule that is not recommended is refused
    When shampoo init runs with the task source "todo.md" and the check "style" running "make style"
    Then the exit code is 2
    And the file ".shampoo/rules/style.toml" does not exist

  @sha:376d43bce73f
  Scenario: A check without a command is refused
    When shampoo init runs with the task source "todo.md" and the check "lint" running ""
    Then the exit code is 2
    And the file ".shampoo/rules/lint.toml" does not exist
