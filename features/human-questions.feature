Feature: Stop taking tasks while too many questions to a human are open
  An agent asks a human, when the policies say so, as a GitHub issue that
  shampoo records, and goes on without waiting. When more of the recorded
  questions are open than the limit, `shampoo next` hands no new task and
  says to wait for a human, with the questions.

  Background:
    Given a git repository with a commit
    And a task source "todo.md"

  @sha:bf8fde14148c
  Scenario: Open questions within the limit let a new task be taken
    Given 3 questions to a human are open
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:4360dbb16053
  Scenario: More open questions than the limit hand no new task
    Given 4 questions to a human are open
    When I run "shampoo next"
    Then the exit code is 0
    And the next action is "wait-for-human"
    And the next action names 4 questions

  @sha:b459ba145035
  Scenario: Answered questions are not counted
    Given 3 questions to a human are open
    And 2 questions to a human are closed
    When I run "shampoo next"
    Then the next action is "take-task"

  @skip
  @sha:a77cb0a77be8
  Scenario: The settings override the limit
    Given settings allowing at most 0 open questions
    And 1 question to a human is open
    When I run "shampoo next"
    Then the next action is "wait-for-human"

  @skip
  @sha:76eba69f6475
  Scenario: Issues shampoo did not open are not counted
    Given 4 issues are open that shampoo did not open
    When I run "shampoo next"
    Then the next action is "take-task"

  @sha:20646b13e53c
  Scenario: A task already taken goes on
    Given 4 questions to a human are open
    And a branch "task-a"
    And no pull request is open
    When I run "shampoo next"
    Then the next action is "continue"
