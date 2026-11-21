import json
import os
import secrets
import shlex
import subprocess
import sys
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, then, when

REPO_ROOT = Path(__file__).resolve().parents[2]

_SHAMPOO_BIN_OPTION = "--shampoo-bin"

VIOLATION = '{ other = "task-failure" }'


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(_SHAMPOO_BIN_OPTION, action="store", default=None)


def pytest_bdd_apply_tag(tag, function):
    return True if tag.startswith("sha:") else None


@pytest.fixture(scope="session")
def shampoo_binary(request, tmp_path_factory):
    bin_option = request.config.getoption(_SHAMPOO_BIN_OPTION)
    if bin_option:
        bin_path = Path(bin_option).resolve()
        assert bin_path.is_file(), f"binary not found: {bin_path}"
        return bin_path
    bin_path = tmp_path_factory.mktemp("bin") / "shampoo"
    subprocess.run(["go", "build", "-o", str(bin_path), "."], cwd=REPO_ROOT, check=True)
    return bin_path


@pytest.fixture
def workspace(tmp_path):
    workspace_dir = tmp_path / "workspace"
    workspace_dir.mkdir()
    return workspace_dir


@pytest.fixture
def run_result():
    return {}


FAKE_GH = """\
import json, sys
from pathlib import Path

state = Path({state!r})
args = sys.argv[1:]
if args[:2] == ["pr", "view"]:
    if not (state / "pr.json").exists():
        print("no pull requests found", file=sys.stderr)
        sys.exit(1)
    print((state / "pr.json").read_text())
elif args[:2] == ["run", "list"]:
    print((state / "runs.json").read_text())
elif args[:2] == ["pr", "merge"]:
    with (state / "merges.log").open("a") as log:
        log.write(" ".join(args) + "\\n")
elif args[:2] == ["issue", "create"]:
    issues = json.loads((state / "issues.json").read_text())
    url = "https://github.example/issues/" + str(len(issues) + 1)
    value = lambda flag: args[args.index(flag) + 1]
    issues.append({{"url": url, "labels": [value("--label")], "state": "OPEN", "title": value("--title"), "body": value("--body")}})
    (state / "issues.json").write_text(json.dumps(issues))
    print(url)
elif args[:2] == ["issue", "list"]:
    label = args[args.index("--label") + 1]
    issues = json.loads((state / "issues.json").read_text())
    print(json.dumps([issue for issue in issues if label in issue["labels"] and issue["state"] == "OPEN"]))
elif args[:2] == ["label", "list"]:
    print(json.dumps([{{"name": name}} for name in json.loads((state / "labels.json").read_text())]))
elif args[:2] == ["label", "create"]:
    labels = json.loads((state / "labels.json").read_text())
    if args[2] in labels:
        print("label already exists", file=sys.stderr)
        sys.exit(1)
    (state / "labels.json").write_text(json.dumps(labels + [args[2]]))
else:
    print("fake gh: unexpected " + " ".join(args), file=sys.stderr)
    sys.exit(1)
"""


@pytest.fixture
def gh_state(tmp_path):
    state = tmp_path / "gh-state"
    state.mkdir()
    (state / "runs.json").write_text("[]")
    (state / "issues.json").write_text("[]")
    (state / "labels.json").write_text("[]")
    return state


@pytest.fixture
def fake_bin(tmp_path, gh_state):
    bin_dir = tmp_path / "fake-bin"
    bin_dir.mkdir()
    gh = bin_dir / "gh"
    gh.write_text(f"#!{sys.executable}\n" + FAKE_GH.format(state=str(gh_state)))
    gh.chmod(0o755)
    return bin_dir


@pytest.fixture
def settings_sections():
    return {}


def toml_value(value) -> str:
    if isinstance(value, dict):
        return "{ " + ", ".join(f"{key} = {toml_value(item)}" for key, item in value.items()) + " }"
    return json.dumps(value)


def write_settings(workspace: Path, sections: dict) -> None:
    (workspace / ".shampoo").mkdir(parents=True, exist_ok=True)
    text = ""
    for section, values in sections.items():
        text += f"[{section}]\n" + "".join(f"{key} = {toml_value(value)}\n" for key, value in values.items())
    (workspace / ".shampoo" / "settings.toml").write_text(text)


def run_shampoo(workspace: Path, shampoo_binary: Path, fake_bin: Path, command: str, stdin: str = ""):
    args = command.split()
    assert args[0] == "shampoo"
    return subprocess.run(
        [str(shampoo_binary), *args[1:]],
        cwd=workspace,
        input=stdin,
        capture_output=True,
        text=True,
        timeout=300,
        env={**os.environ, "PATH": f"{fake_bin}{os.pathsep}{os.environ['PATH']}"},
    )


@given("a git repository with a commit", target_fixture="default_branch")
def given_git_repository(workspace):
    git = ["git", "-C", str(workspace)]
    subprocess.run([*git, "init", "-q"], check=True)
    subprocess.run([*git, "config", "user.name", "acceptance"], check=True)
    subprocess.run([*git, "config", "user.email", "acceptance@example.com"], check=True)
    (workspace / "README.md").write_text(secrets.token_hex(8) + "\n")
    subprocess.run([*git, "add", "-A"], check=True)
    subprocess.run([*git, "commit", "-q", "-m", "init"], check=True)
    head = subprocess.run([*git, "rev-parse", "--abbrev-ref", "HEAD"], check=True, capture_output=True, text=True)
    return head.stdout.strip()


@given(parsers.parse('a changed file "{path}"'))
def given_changed_file(workspace, path):
    file = workspace / path
    file.parent.mkdir(parents=True, exist_ok=True)
    with file.open("a") as f:
        f.write(secrets.token_hex(8) + "\n")


@given(parsers.parse('"{command}" has passed'))
def given_command_passed(workspace, shampoo_binary, fake_bin, command):
    result = run_shampoo(workspace, shampoo_binary, fake_bin, command)
    assert result.returncode == 0, (
        f"precondition failed: {command} exited {result.returncode}: "
        f"stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@given(parsers.parse('"{command}" has run'))
def given_command_ran(workspace, shampoo_binary, fake_bin, command):
    run_shampoo(workspace, shampoo_binary, fake_bin, command)


@given(parsers.parse('a branch "{name}"'))
def given_branch(workspace, name):
    subprocess.run(["git", "-C", str(workspace), "checkout", "-q", "-b", name], check=True)


@given("the changes are committed")
def given_changes_committed(workspace):
    git = ["git", "-C", str(workspace)]
    subprocess.run([*git, "add", "-A"], check=True)
    subprocess.run([*git, "commit", "-q", "-m", "changes"], check=True)


@given(parsers.parse('the branch "{name}" is committed and merged'))
def given_branch_merged(workspace, default_branch, name):
    git = ["git", "-C", str(workspace)]
    subprocess.run([*git, "add", "-A"], check=True)
    subprocess.run([*git, "commit", "-q", "-m", name], check=True)
    subprocess.run([*git, "checkout", "-q", default_branch], check=True)
    subprocess.run([*git, "merge", "-q", "--no-ff", "-m", f"Merge {name}", name], check=True)


@when(parsers.parse('I run "{command}"'))
def when_running(workspace, shampoo_binary, fake_bin, command, run_result):
    result = run_shampoo(workspace, shampoo_binary, fake_bin, command)
    run_result.update(
        {
            "stdout": result.stdout,
            "stderr": result.stderr,
            "returncode": result.returncode,
        }
    )


@then(parsers.parse("the exit code is {code:d}"))
def then_exit_code(run_result, code):
    assert run_result["returncode"] == code, (
        f"expected exit {code}, got {run_result['returncode']}: "
        f"stdout={run_result['stdout']!r}, stderr={run_result['stderr']!r}"
    )


@then(parsers.parse('stderr contains "{text}"'))
def then_stderr_contains(run_result, text):
    assert text in run_result["stderr"], (
        f"expected {text!r} in stderr, got {run_result['stderr']!r}"
    )


def write_check(workspace: Path, check_id: str, command: str, needs: str = "") -> None:
    checks = workspace / ".shampoo" / "checks"
    checks.mkdir(parents=True, exist_ok=True)
    (checks / f"{check_id}.toml").write_text(f'command = {command!r}\nsize = "small"\n{needs}')


def write_rule(
    workspace: Path,
    rule_id: str,
    transition: str,
    check_id: str,
    judge: str = f"{{ exit = {VIOLATION} }}",
    budget: str | None = "{ violations = 5, runs = 10 }",
    extra: str = "",
    tables: str = "",
) -> None:
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    transitions = ", ".join(f'"{name}"' for name in transition.split(" and "))
    (rules / f"{rule_id}.toml").write_text(
        f'rule = "{rule_id} holds."\n'
        f'check = "{check_id}"\n'
        f"transitions = [{transitions}]\n"
        f"{extra}"
        f"judge = {judge}\n"
        + (f"budget = {budget}\n" if budget is not None else "")
        + tables
    )


def new_id() -> str:
    return "rule-" + secrets.token_hex(4)


def record(runs_log: Path, check_id: str) -> str:
    return f"echo {shlex.quote(check_id)} >> {shlex.quote(str(runs_log))}"


@given(parsers.parse('a check "{check_id}" that exits {code:d}'))
def given_check_exits(workspace, check_id, code):
    write_check(workspace, check_id, f"exit {code}")


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}"'))
def given_rule(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id)


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" judging exit {code:d} as a workflow failure'))
def given_rule_workflow_failure(workspace, transition, check_id, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        check_id,
        judge=f'{{ exit = {{ {code} = "workflow-failure", other = "task-failure" }} }}',
    )


@given(parsers.parse("settings allowing {count:d} failed attempt in harness repair"))
def given_repair_limit(workspace, settings_sections, count):
    settings_sections.setdefault("state-machine.harness-repair", {})["max-failed-attempts"] = count
    write_settings(workspace, settings_sections)


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" and a zero budget'))
def given_rule_zero_budget(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id, budget="{ violations = 0, runs = 1 }")


@given(parsers.parse('settings asking a human with the label "{label}" and at most {limit:d} open questions'))
def given_human_review_settings(workspace, settings_sections, label, limit):
    settings_sections["human-review"] = {"label": label, "max-open": limit}
    write_settings(workspace, settings_sections)


@given(parsers.parse('a task source "{path}"'))
def given_task_source(workspace, settings_sections, path):
    (workspace / path).write_text("- a task\n")
    settings_sections.setdefault("orchestrator", {})["tasks"] = path
    write_settings(workspace, settings_sections)


def next_action(run_result) -> dict:
    try:
        return json.loads(run_result["stdout"])
    except json.JSONDecodeError as error:
        raise AssertionError(f"expected a JSON object, got {run_result['stdout']!r}: {run_result['stderr']!r}") from error


@then(parsers.parse('the next action is "{action}"'))
def then_next_action(run_result, action):
    got = next_action(run_result).get("action")
    assert got == action, f"expected action {action!r}, got {got!r}: stderr={run_result['stderr']!r}"


@then(parsers.parse('the next action has "{field}" "{value}"'))
def then_next_action_field(run_result, field, value):
    got = next_action(run_result).get(field)
    assert got == value, f"expected {field} {value!r}, got {got!r}"


@given("no pull request is open")
def given_no_pull_request(gh_state):
    (gh_state / "pr.json").unlink(missing_ok=True)


def write_policy(workspace: Path, policy_id: str, review: bool) -> None:
    policies = workspace / ".shampoo" / "policies"
    policies.mkdir(parents=True, exist_ok=True)
    (policies / f"{policy_id}.toml").write_text(
        f'text = "{policy_id} is followed."\n' + ("review = true\n" if review else "")
    )


@given(parsers.parse('a policy "{policy_id}"'))
def given_policy(workspace, policy_id):
    write_policy(workspace, policy_id, review=False)


@given(parsers.parse('a policy "{policy_id}" for review'))
def given_policy_for_review(workspace, policy_id):
    write_policy(workspace, policy_id, review=True)


@given(parsers.parse('a policy "{policy_id}" with the example "{example}"'))
def given_policy_with_example(workspace, policy_id, example):
    write_policy(workspace, policy_id, review=False)
    with (workspace / ".shampoo" / "policies" / f"{policy_id}.toml").open("a") as f:
        f.write(f"example = {json.dumps(example)}\n")


@given(parsers.parse('a policy "{policy_id}" without a text'))
def given_policy_without_text(workspace, policy_id):
    policies = workspace / ".shampoo" / "policies"
    policies.mkdir(parents=True, exist_ok=True)
    (policies / f"{policy_id}.toml").write_text("review = true\n")


@then(parsers.parse('stdout contains "{text}"'))
def then_stdout_contains(run_result, text):
    assert text in run_result["stdout"], f"expected {text!r} in stdout, got {run_result['stdout']!r}"


@then(parsers.parse('stdout does not contain "{text}"'))
def then_stdout_lacks(run_result, text):
    assert text not in run_result["stdout"], f"stdout={run_result['stdout']!r}"


@then("stdout is empty")
def then_stdout_empty(run_result):
    assert run_result["stdout"] == "", f"expected no stdout, got {run_result['stdout']!r}"
