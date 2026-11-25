Feature: Help and version
  `shampoo help` prints what each command does, and `shampoo version`
  prints the version of the build, each to stdout with exit 0. A release
  build carries its tag as its version; any other build prints `dev`.

  @sha:2c8f7cf2c9d3
  Scenario Outline: Help is printed for each of its spellings
    When I run "shampoo <argument>"
    Then the exit code is 0
    And stdout contains "shampoo guard <transition>"
    And stdout contains "shampoo hook <transition>"
    And stdout contains "shampoo next"

    Examples:
      | argument |
      | help     |
      | -h       |
      | --help   |

  @sha:6c2668478f8b
  Scenario Outline: The version of a release build is printed for each of its spellings
    Given shampoo built as version "1.2.3"
    When I run that shampoo with "<argument>"
    Then the exit code is 0
    And stdout is "1.2.3"

    Examples:
      | argument  |
      | version   |
      | -v        |
      | --version |

  @sha:0d4fc8e1bcc5
  Scenario: A build that is not a release prints dev as its version
    When I run "shampoo version"
    Then the exit code is 0
    And stdout is "dev"
