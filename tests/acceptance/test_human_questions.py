from pytest_bdd import given, parsers, scenarios, then

from conftest import next_action, open_issue

scenarios("../../features/human-questions.feature")


@given(parsers.re(r"(?P<count>\d+) questions? to a human (?:is|are) open"))
def given_open_questions(workspace, gh_state, count):
    for _ in range(int(count)):
        open_issue(workspace, gh_state, "human-review", "question", recorded=True)


@given(parsers.re(r"(?P<count>\d+) questions? to a human (?:is|are) closed"))
def given_closed_questions(workspace, gh_state, count):
    for _ in range(int(count)):
        open_issue(workspace, gh_state, "human-review", "question", recorded=True, state="CLOSED")


@given(parsers.parse("{count:d} issues are open that shampoo did not open"))
def given_foreign_issues(workspace, gh_state, count):
    for _ in range(count):
        open_issue(workspace, gh_state, "human-review", "question", recorded=False)


@then(parsers.parse("the next action names {count:d} questions"))
def then_questions(run_result, count):
    got = next_action(run_result).get("questions", [])
    assert len(got) == count, f"got questions {got!r}"
