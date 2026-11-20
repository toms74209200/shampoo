import json
import secrets
import subprocess
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

from conftest import next_action, write_settings

scenarios("../../features/pull-request.feature")

FAKE_GH = """\
import json, sys
from pathlib import Path

state = Path({state!r})
args = sys.argv[1:]
if args[:2] == ["pr", "view"]:
    print((state / "pr.json").read_text())
elif args[:2] == ["run", "list"]:
    print((state / "runs.json").read_text())
elif args[:2] == ["pr", "merge"]:
    with (state / "merges.log").open("a") as log:
        log.write(" ".join(args) + "\\n")
else:
    print("fake gh: unexpected " + " ".join(args), file=sys.stderr)
    sys.exit(1)
"""


@pytest.fixture
def gh_state(tmp_path):
    state = tmp_path / "gh-state"
    state.mkdir()
    (state / "runs.json").write_text("[]")
    return state


def read_json(path: Path):
    return json.loads(path.read_text())


def write_workflow(workspace: Path, name: str, filters: str) -> None:
    workflows = workspace / ".github" / "workflows"
    workflows.mkdir(parents=True, exist_ok=True)
    (workflows / f"{name}.yml").write_text(
        f"name: {name}\n"
        "on:\n"
        "  pull_request:\n"
        f"{filters}"
        "jobs:\n"
        "  check:\n"
        "    runs-on: ubuntu-latest\n"
        "    steps:\n"
        "      - run: 'true'\n"
    )


@given(parsers.parse('a workflow "{name}" on pull requests to paths "{paths}"'))
def given_workflow_paths(workspace, name, paths):
    write_workflow(workspace, name, f"    paths:\n      - '{paths}'\n")


@given(parsers.parse('a workflow "{name}" on pull requests'))
def given_workflow(workspace, name):
    write_workflow(workspace, name, "")


@given(parsers.parse('settings merging by "{method}"'))
def given_merge_method(workspace, settings_sections, method):
    settings_sections.setdefault("orchestrator", {})["merge"] = method
    write_settings(workspace, settings_sections)


@given(parsers.parse("settings waiting at most {minutes:d} minutes for checks"))
def given_max_wait(workspace, settings_sections, minutes):
    settings_sections.setdefault("orchestrator", {})["max-wait-minutes"] = minutes
    write_settings(workspace, settings_sections)


@given("a pull request is open for the branch")
def given_pull_request(workspace, default_branch, fake_bin, gh_state):
    git = ["git", "-C", str(workspace)]
    log = subprocess.run(
        [*git, "log", "--reverse", "--format=%H %cI", f"{default_branch}..HEAD"],
        check=True, capture_output=True, text=True,
    ).stdout.split()
    files = subprocess.run(
        [*git, "diff", "--name-only", f"{default_branch}...HEAD"],
        check=True, capture_output=True, text=True,
    ).stdout.split()
    commits = [{"oid": oid, "committedDate": date} for oid, date in zip(log[::2], log[1::2])]
    (gh_state / "pr.json").write_text(json.dumps({
        "number": 1,
        "state": "OPEN",
        "baseRefName": default_branch,
        "headRefOid": commits[-1]["oid"],
        "commits": commits,
        "files": [{"path": path} for path in files],
    }))
    gh = fake_bin / "gh"
    gh.write_text(f"#!{sys.executable}\n" + FAKE_GH.format(state=str(gh_state)))
    gh.chmod(0o755)


def add_run(gh_state: Path, workflow: str, status: str, conclusion: str) -> None:
    pull_request = read_json(gh_state / "pr.json")
    runs = read_json(gh_state / "runs.json")
    run_id = len(runs) + 1
    created = datetime.now(timezone.utc) + timedelta(seconds=run_id)
    runs.append({
        "databaseId": run_id,
        "workflowName": workflow,
        "headSha": pull_request["commits"][-1]["oid"],
        "status": status,
        "conclusion": conclusion,
        "createdAt": created.isoformat(timespec="seconds").replace("+00:00", "Z"),
        "url": f"https://github.example/runs/{run_id}-{secrets.token_hex(4)}",
    })
    (gh_state / "runs.json").write_text(json.dumps(runs))


@given(parsers.parse('the workflow "{workflow}" is running on the latest commit'))
def given_run_in_progress(gh_state, workflow):
    add_run(gh_state, workflow, "in_progress", "")


@given(parsers.parse('the workflow "{workflow}" has succeeded on the latest commit'))
def given_run_succeeded(gh_state, workflow):
    add_run(gh_state, workflow, "completed", "success")


@given(parsers.parse('the workflow "{workflow}" has failed on the latest commit'))
def given_run_failed(gh_state, workflow):
    add_run(gh_state, workflow, "completed", "failure")


@given(parsers.parse('the workflow "{workflow}" did not execute on the latest commit'))
def given_run_not_executed(gh_state, workflow):
    add_run(gh_state, workflow, "completed", "action_required")


@given("the run pushed a commit to the pull request")
def given_pushed_by_run(gh_state):
    pull_request = read_json(gh_state / "pr.json")
    oid = secrets.token_hex(20)
    pull_request["commits"].append({
        "oid": oid,
        "committedDate": datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z"),
    })
    pull_request["headRefOid"] = oid
    pull_request["files"].append({"path": f"features/{secrets.token_hex(4)}.feature"})
    (gh_state / "pr.json").write_text(json.dumps(pull_request))


@then(parsers.parse('the next action names the failed workflow "{workflow}" with its run\'s URL'))
def then_failed_workflow(run_result, gh_state, workflow):
    runs = [run for run in read_json(gh_state / "runs.json") if run["workflowName"] == workflow]
    failures = next_action(run_result).get("failures")
    assert failures == [{"workflow": workflow, "url": runs[-1]["url"]}], f"got failures {failures!r}"


@then(parsers.parse('the pull request is merged with "{flag}"'))
def then_merged_with(gh_state, flag):
    merges = gh_state / "merges.log"
    assert merges.exists(), "the pull request was not merged"
    assert flag in merges.read_text().split(), f"expected {flag!r} in {merges.read_text()!r}"


@then("the pull request is not merged")
def then_not_merged(gh_state):
    merges = gh_state / "merges.log"
    assert not merges.exists(), f"the pull request was merged: {merges.read_text()!r}"
