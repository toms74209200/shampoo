import json
import tomllib

from pytest_bdd import given, parsers, scenarios, then, when

from conftest import run_shampoo

scenarios("../../features/init.feature")

HAND_WRITTEN = 'unchecked = "written by hand"\n'


def run_init(workspace, shampoo_binary, fake_bin, run_result, checks: dict) -> None:
    stdin = json.dumps({"tasks": "todo.md", "checks": checks})
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo init", stdin=stdin)
    run_result.update({"stdout": result.stdout, "stderr": result.stderr, "returncode": result.returncode})


@when(parsers.re(r'shampoo init runs with the task source "todo.md" and the check "(?P<check>[^"]+)" running "(?P<command>[^"]*)"$'))
def when_init(workspace, shampoo_binary, fake_bin, run_result, check, command):
    run_init(workspace, shampoo_binary, fake_bin, run_result, {check: {"command": command}})


@when(parsers.re(r'shampoo init runs with the task source "todo.md" and the check "(?P<check>[^"]+)" running "(?P<command>[^"]*)" writing its exceptions to "(?P<file>[^"]+)"$'))
def when_init_exceptions(workspace, shampoo_binary, fake_bin, run_result, check, command, file):
    run_init(workspace, shampoo_binary, fake_bin, run_result, {check: {"command": command, "exceptions": file}})


@given(parsers.parse('a task source file "{path}"'))
def given_task_source_file(workspace, path):
    (workspace / path).write_text("- a task\n")


@given(parsers.parse('a rule "{rule_id}" written by hand'))
def given_hand_written_rule(workspace, rule_id):
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    (rules / f"{rule_id}.toml").write_text(HAND_WRITTEN)


@given(parsers.parse('a label "{name}" exists'))
def given_label(gh_state, name):
    labels = json.loads((gh_state / "labels.json").read_text())
    (gh_state / "labels.json").write_text(json.dumps(labels + [name]))


@then(parsers.parse('the file "{path}" exists'))
def then_file_exists(workspace, path):
    assert (workspace / path).is_file(), f"{path} does not exist"


@then(parsers.parse('the file "{path}" does not exist'))
def then_file_missing(workspace, path):
    assert not (workspace / path).exists(), f"{path} exists"


@then(parsers.parse('the check "{check_id}" runs "{command}"'))
def then_check_command(workspace, check_id, command):
    check = tomllib.loads((workspace / ".shampoo" / "checks" / f"{check_id}.toml").read_text())
    assert check["command"] == command, f"got {check!r}"


@then(parsers.parse('the settings have the task source "{path}"'))
def then_settings_tasks(workspace, path):
    settings = tomllib.loads((workspace / ".shampoo" / "settings.toml").read_text())
    assert settings["orchestrator"]["tasks"] == path, f"got {settings!r}"


@then(parsers.parse('the rule "{rule_id}" declares its exceptions in "{file}"'))
def then_rule_exceptions(workspace, rule_id, file):
    rule = tomllib.loads((workspace / ".shampoo" / "rules" / f"{rule_id}.toml").read_text())
    assert rule["exceptions"]["file"] == file, f"got {rule!r}"


@then(parsers.parse('the rule "{rule_id}" is the one written by hand'))
def then_rule_kept(workspace, rule_id):
    text = (workspace / ".shampoo" / "rules" / f"{rule_id}.toml").read_text()
    assert text == HAND_WRITTEN, f"got {text!r}"


@then(parsers.parse('init reports "{path}" as kept'))
def then_reported_kept(run_result, path):
    report = json.loads(run_result["stdout"])
    assert path in report["kept"], f"got {report!r}"


@then(parsers.parse('the labels are "{first}" and "{second}"'))
def then_labels(gh_state, first, second):
    labels = json.loads((gh_state / "labels.json").read_text())
    assert sorted(labels) == sorted([first, second]), f"got labels {labels!r}"


def claude_settings(workspace) -> dict:
    return json.loads((workspace / ".claude" / "settings.json").read_text())


@given(parsers.parse('Claude Code settings with the permission "{permission}"'))
def given_claude_settings(workspace, permission):
    (workspace / ".claude").mkdir(parents=True, exist_ok=True)
    (workspace / ".claude" / "settings.json").write_text(json.dumps({"permissions": {"allow": [permission]}}))


@then(parsers.parse('the policy "{policy_id}" is marked for review'))
def then_policy_review(workspace, policy_id):
    policy = tomllib.loads((workspace / ".shampoo" / "policies" / f"{policy_id}.toml").read_text())
    assert policy.get("review") is True, f"got {policy!r}"


@then(parsers.parse('the Claude Code settings have a "{event}" hook running "{command}"'))
def then_hook(workspace, event, command):
    commands = [hook["command"] for entry in claude_settings(workspace)["hooks"][event] for hook in entry["hooks"]]
    assert any(command in got for got in commands), f"got {commands!r}"


@then(parsers.parse('the Claude Code settings allow "{permission}"'))
def then_permission(workspace, permission):
    allowed = claude_settings(workspace)["permissions"]["allow"]
    assert permission in allowed, f"got {allowed!r}"
