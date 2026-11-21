import json
import subprocess
from pathlib import Path

from pytest_bdd import given, parsers, scenarios, then

from conftest import new_id, next_action, run_shampoo, write_settings

scenarios("../../features/workflow-review.feature")


def stats_path(workspace: Path) -> Path:
    common = subprocess.run(
        ["git", "-C", str(workspace), "rev-parse", "--path-format=absolute", "--git-common-dir"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()
    return Path(common) / "shampoo" / "stats.jsonl"


def add_merged_tasks(workspace: Path, count: int, questions: int, repairs: int) -> None:
    stats = stats_path(workspace)
    stats.parent.mkdir(parents=True, exist_ok=True)
    with stats.open("a") as file:
        for _ in range(count):
            summary = {"task": new_id(), "attempts": {}, "violations": {}, "findings": {}, "millis": {}, "repairs": repairs, "questions": questions}
            file.write(json.dumps(summary) + "\n")


def issues(gh_state) -> list:
    return json.loads((gh_state / "issues.json").read_text())


@given(parsers.re(r"(?P<count>\d+) merged tasks? that asked a human"))
def given_tasks_asking(workspace, count):
    add_merged_tasks(workspace, int(count), questions=1, repairs=0)


@given(parsers.re(r"(?P<count>\d+) merged tasks? that entered harness repair"))
def given_tasks_repairing(workspace, count):
    add_merged_tasks(workspace, int(count), questions=0, repairs=1)


@given(parsers.re(r"(?P<count>\d+) merged tasks? that did not"))
def given_tasks_plain(workspace, count):
    add_merged_tasks(workspace, int(count), questions=0, repairs=0)


@given(parsers.parse("settings reviewing the workflow over {tasks:d} tasks with at most {questions:d} questions and {repairs:d} repairs"))
def given_review_settings(workspace, settings_sections, tasks, questions, repairs):
    settings_sections.setdefault("state-machine.workflow-review", {}).update({"tasks": tasks, "max-questions": questions, "max-repairs": repairs})
    write_settings(workspace, settings_sections)


@given(parsers.parse("settings allowing {count:d} failed workflow reviews"))
def given_review_attempts(workspace, settings_sections, count):
    settings_sections.setdefault("state-machine.workflow-review", {})["max-failed-attempts"] = count
    write_settings(workspace, settings_sections)


@given('"shampoo next" has said to review the workflow')
def given_reviewed(workspace, shampoo_binary, fake_bin):
    add_merged_tasks(workspace, 11, questions=1, repairs=0)
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo next")
    assert result.returncode == 0 and json.loads(result.stdout)["action"] == "review-workflow", (
        f"precondition failed: shampoo next exited {result.returncode}: stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@given('"shampoo next" has handed the workflow to a human')
def given_handed(workspace, shampoo_binary, fake_bin):
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo next")
    assert result.returncode == 0 and json.loads(result.stdout)["action"] == "wait-for-human", (
        f"precondition failed: shampoo next exited {result.returncode}: stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@given(parsers.parse('1 issue labeled "{label}" is open'))
def given_open_issue(gh_state, label):
    opened = issues(gh_state) + [{"url": f"https://github.example/issues/{new_id()}", "labels": [label], "state": "OPEN"}]
    (gh_state / "issues.json").write_text(json.dumps(opened))


@given(parsers.parse('the issue labeled "{label}" is closed'))
def given_closed_issue(gh_state, label):
    closed = [{**issue, "state": "CLOSED"} if label in issue["labels"] else issue for issue in issues(gh_state)]
    (gh_state / "issues.json").write_text(json.dumps(closed))


@then(parsers.parse('an issue labeled "{label}" is open'))
def then_issue_open(gh_state, label):
    got = [issue for issue in issues(gh_state) if label in issue["labels"] and issue["state"] == "OPEN"]
    assert len(got) == 1, f"got issues {issues(gh_state)!r}"


@then(parsers.parse("the next action names {count:d} question"))
def then_one_question(run_result, count):
    got = next_action(run_result).get("questions", [])
    assert len(got) == count, f"got questions {got!r}"


@then(parsers.parse("the next action names {count:d} tasks that asked a human"))
def then_asked(run_result, count):
    got = next_action(run_result).get("asked", [])
    assert len(got) == count, f"got asked {got!r}"


@then(parsers.parse("the next action names {count:d} tasks that repaired the harness"))
def then_repaired(run_result, count):
    got = next_action(run_result).get("repaired", [])
    assert len(got) == count, f"got repaired {got!r}"
