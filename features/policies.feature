Feature: Print the policies for the agent's context
  `shampoo policies` prints every policy of the project, its text and its
  code example, for the hook that puts them into the agent's context when
  a session starts and after the context is compacted.

  Background:
    Given a git repository with a commit

  @sha:7fe6f62c0f21
  Scenario: Every policy is printed with its text
    Given a policy "layout"
    And a policy "functions" for review
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout contains "layout is followed."
    And stdout contains "functions is followed."

  @sha:2f6aea638930
  Scenario: A policy's code example is printed with it
    Given a policy "layout" with the example "main.go only runs the program"
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout contains "main.go only runs the program"

  @sha:1a93fd7c8c5c
  Scenario: Without policies only asking a human is printed
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout contains "shampoo ask"
    And stdout does not contain "Policies of this project"

  @sha:297543a7d590
  Scenario: Asking a human follows the policies
    Given a policy "layout"
    When I run "shampoo policies"
    Then stdout contains "layout is followed."
    And stdout contains "shampoo ask"

  @sha:6a533c61df95
  Scenario: An invalid policy is a workflow failure
    Given a policy "layout" without a text
    When I run "shampoo policies"
    Then the exit code is 2
