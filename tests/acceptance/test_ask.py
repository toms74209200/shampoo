import json
import subprocess
from pathlib import Path

from pytest_bdd import given, parsers, scenarios, then, when

from conftest import run_shampoo

scenarios("../../features/ask.feature")


def question(decision: str, alternatives: str, tradeoffs: str) -> str:
    return json.dumps({"decision": decision, "alternatives": alternatives, "tradeoffs": tradeoffs})


def issues(gh_state) -> list:
    return json.loads((gh_state / "issues.json").read_text())


@when(parsers.parse('I run "shampoo ask" deciding "{decision}" instead of "{alternatives}" for "{tradeoffs}"'))
def when_asking(workspace, shampoo_binary, fake_bin, run_result, decision, alternatives, tradeoffs):
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo ask", stdin=question(decision, alternatives, tradeoffs))
    run_result.update({"stdout": result.stdout, "stderr": result.stderr, "returncode": result.returncode})


@given(parsers.parse('"shampoo ask" has asked "{decision}"'))
def given_asked(workspace, shampoo_binary, fake_bin, decision):
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo ask", stdin=question(decision, "another way", "a trade-off"))
    assert result.returncode == 0, (
        f"precondition failed: shampoo ask exited {result.returncode}: stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@then(parsers.parse('an issue labeled "{label}" is open titled "{title}"'))
def then_issue(gh_state, label, title):
    got = [(issue["labels"], issue["state"], issue["title"]) for issue in issues(gh_state)]
    assert got == [([label], "OPEN", title)], f"got issues {got!r}"


@then(parsers.parse('the issue says "{text}"'))
def then_issue_says(gh_state, text):
    body = issues(gh_state)[0]["body"]
    assert text in body, f"body={body!r}"


@then("no issue is open")
def then_no_issue(gh_state):
    assert issues(gh_state) == [], f"got issues {issues(gh_state)!r}"


@then(parsers.parse('the statistics have {count:d} question for task "{task}"'))
def then_statistics_questions(workspace, count, task):
    common = subprocess.run(
        ["git", "-C", str(workspace), "rev-parse", "--path-format=absolute", "--git-common-dir"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()
    stats = Path(common) / "shampoo" / "stats.jsonl"
    lines = [json.loads(line) for line in stats.read_text().splitlines()] if stats.exists() else []
    got = [line.get("questions") for line in lines if line.get("task") == task]
    assert got == [count], f"got questions {got!r} in {lines!r}"
