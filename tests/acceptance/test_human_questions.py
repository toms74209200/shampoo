import json

from pytest_bdd import given, parsers, scenarios, then

from conftest import next_action

scenarios("../../features/human-questions.feature")


def add_questions(gh_state, count: int, label: str, state: str) -> None:
    issues = json.loads((gh_state / "issues.json").read_text())
    for _ in range(count):
        number = len(issues) + 1
        issues.append({"url": f"https://github.example/issues/{number}", "labels": [label], "state": state})
    (gh_state / "issues.json").write_text(json.dumps(issues))


@given(parsers.re(r"(?P<count>\d+) questions? to a human (?:is|are) open"))
def given_open_questions(gh_state, count):
    add_questions(gh_state, int(count), "human-review", "OPEN")


@given(parsers.re(r"(?P<count>\d+) questions? to a human (?:is|are) closed"))
def given_closed_questions(gh_state, count):
    add_questions(gh_state, int(count), "human-review", "CLOSED")


@given(parsers.re(r'(?P<count>\d+) questions? to a human labeled "(?P<label>[^"]+)" (?:is|are) open'))
def given_labeled_questions(gh_state, count, label):
    add_questions(gh_state, int(count), label, "OPEN")


@then(parsers.parse("the next action names {count:d} questions"))
def then_questions(run_result, count):
    got = next_action(run_result).get("questions", [])
    assert len(got) == count, f"got questions {got!r}"
