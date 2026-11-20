Feature: Print the policies for the agent's context
  `shampoo policies` prints every policy of the project, its text and its
  code example, for the hook that puts them into the agent's context when
  a session starts and after the context is compacted.

  Background:
    Given a git repository with a commit

  @skip
  Scenario: Every policy is printed with its text
    Given a policy "layout"
    And a policy "functions" for review
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout contains "layout is followed."
    And stdout contains "functions is followed."

  @skip
  Scenario: A policy's code example is printed with it
    Given a policy "layout" with the example "main.go only runs the program"
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout contains "main.go only runs the program"

  @skip
  Scenario: Without policies nothing is printed
    When I run "shampoo policies"
    Then the exit code is 0
    And stdout is empty

  @skip
  Scenario: An invalid policy is a workflow failure
    Given a policy "layout" without a text
    When I run "shampoo policies"
    Then the exit code is 2
