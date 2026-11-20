import json

from pytest_bdd import given, parsers, scenarios, then

from conftest import write_settings

scenarios("../../features/next.feature")


@given(parsers.parse('a task source "{path}"'))
def given_task_source(workspace, settings_sections, path):
    (workspace / path).write_text("- a task\n")
    settings_sections.setdefault("orchestrator", {})["tasks"] = path
    write_settings(workspace, settings_sections)


@given(parsers.parse('a task source "{path}" that does not exist'))
def given_missing_task_source(workspace, settings_sections, path):
    settings_sections.setdefault("orchestrator", {})["tasks"] = path
    write_settings(workspace, settings_sections)


@given("no pull request is open")
def given_no_pull_request(fake_bin):
    gh = fake_bin / "gh"
    gh.write_text("#!/bin/sh\necho 'no pull requests found' >&2\nexit 1\n")
    gh.chmod(0o755)


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
