Feature: Stop taking tasks while too many questions to a human are open
  An agent asks a human, when the policies say so, as a GitHub issue with
  the label for questions, and goes on without waiting. When more
  questions are open than the limit, `shampoo next` hands no new task and
  says to wait for a human, with the questions.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"

  @skip
  Scenario: Open questions within the limit let a new task be taken
    Given 3 questions to a human are open
    When I run "shampoo next"
    Then the next action is "take-task"

  @skip
  Scenario: More open questions than the limit hand no new task
    Given 4 questions to a human are open
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "wait-for-human"
    And the next action names 4 questions

  @skip
  Scenario: Answered questions are not counted
    Given 3 questions to a human are open
    And 2 questions to a human are closed
    When I run "shampoo next"
    Then the next action is "take-task"

  @skip
  Scenario: The settings override the label and the limit
    Given settings asking a human with the label "ask" and at most 0 open questions
    And 1 question to a human labeled "ask" is open
    When I run "shampoo next"
    Then the next action is "wait-for-human"

  @skip
  Scenario: A task already taken goes on
    Given 4 questions to a human are open
    And a branch "task-a"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "continue"
