import secrets
import shlex
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

scenarios("../../features/guard.feature")


@pytest.fixture
def runs_log(tmp_path):
    return tmp_path / "runs.log"


def write_check(workspace: Path, check_id: str, command: str, exit_map: str, needs: str = "") -> None:
    checks = workspace / ".shampoo" / "checks"
    checks.mkdir(parents=True, exist_ok=True)
    (checks / f"{check_id}.toml").write_text(
        f"command = {command!r}\n"
        f'size = "small"\n'
        f"exit = {exit_map}\n"
        f"{needs}"
    )


def write_rule(workspace: Path, rule_id: str, transition: str, check_id: str, extra: str = "") -> None:
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    (rules / f"{rule_id}.toml").write_text(
        f'rule = "{rule_id} holds."\n'
        f'check = "{check_id}"\n'
        f'transitions = ["{transition}"]\n'
        f"{extra}"
    )


def new_id() -> str:
    return "rule-" + secrets.token_hex(4)


def record(runs_log: Path, check_id: str) -> str:
    return f"echo {shlex.quote(check_id)} >> {shlex.quote(str(runs_log))}"


@given(parsers.parse('a check "{check_id}" that exits {code:d}'))
def given_check_exits(workspace, check_id, code):
    write_check(workspace, check_id, f"exit {code}", '{ other = "task-failure" }')


@given(parsers.parse('a check "{check_id}" that prints "{text}" and exits {code:d} as a task failure'))
def given_check_prints(workspace, check_id, text, code):
    write_check(
        workspace,
        check_id,
        f"echo {shlex.quote(text)}; exit {code}",
        f'{{ {code} = "task-failure", other = "workflow-failure" }}',
    )


@given(parsers.parse('a check "{check_id}" that exits {code:d} as a task failure'))
def given_check_task_failure(workspace, check_id, code):
    write_check(workspace, check_id, f"exit {code}", f'{{ {code} = "task-failure", other = "workflow-failure" }}')


@given(parsers.parse('a check "{check_id}" that exits {code:d} as a workflow failure'))
def given_check_workflow_failure(workspace, check_id, code):
    write_check(workspace, check_id, f"exit {code}", f'{{ {code} = "workflow-failure", other = "task-failure" }}')


@given(parsers.parse('a check "{check_id}" that records that it ran'))
def given_check_records(workspace, runs_log, check_id):
    write_check(workspace, check_id, record(runs_log, check_id), '{ other = "task-failure" }')


@given(parsers.parse('a check "{check_id}" that needs "{needed}" and records that it ran'))
def given_check_needs(workspace, runs_log, check_id, needed):
    write_check(
        workspace,
        check_id,
        record(runs_log, check_id),
        '{ other = "task-failure" }',
        f'needs = ["{needed}"]\n',
    )


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}"'))
def given_rule(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id)


@given(parsers.parse('a rule "{rule_id}" guarding {transition:w} with check "{check_id}"'))
def given_named_rule(workspace, rule_id, transition, check_id):
    write_rule(workspace, rule_id, transition, check_id)


@given(parsers.parse('a rule guarding {transition:w} for paths "{paths}" with check "{check_id}"'))
def given_rule_paths(workspace, transition, paths, check_id):
    write_rule(workspace, new_id(), transition, check_id, f'paths = ["{paths}"]\n')


@given(
    parsers.parse(
        'a rule "{rule_id}" guarding {transition:w} with check "{check_id}" whose judge prints "{text}" and exits {code:d} as a task failure'
    )
)
def given_rule_judge_fails(workspace, rule_id, transition, check_id, text, code):
    command = f"echo {shlex.quote(text)}; exit {code}"
    write_rule(
        workspace,
        rule_id,
        transition,
        check_id,
        f'judge = {{ command = {command!r}, exit = {{ {code} = "task-failure", other = "workflow-failure" }} }}\n',
    )


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" whose judge exits {code:d} charging its budget'))
def given_rule_judge_budget(workspace, transition, check_id, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        check_id,
        f'judge = {{ command = "exit {code}", exit = {{ {code} = "budget", other = "workflow-failure" }} }}\nbudget = 0.1\n',
    )


def runs(runs_log: Path) -> list[str]:
    return runs_log.read_text().splitlines() if runs_log.exists() else []


@then(parsers.parse('the check "{check_id}" has run {times}'))
def then_check_ran(runs_log, check_id, times):
    expected = {"once": 1, "twice": 2}[times]
    count = runs(runs_log).count(check_id)
    assert count == expected, f"expected {check_id} to run {expected} times, got {count}"


@then(parsers.parse('the check "{check_id}" has not run'))
def then_check_not_run(runs_log, check_id):
    assert check_id not in runs(runs_log), f"expected {check_id} not to run, got {runs(runs_log)}"


@then(parsers.parse('the checks ran in the order "{first}", "{second}"'))
def then_checks_order(runs_log, first, second):
    assert runs(runs_log) == [first, second], f"expected {[first, second]}, got {runs(runs_log)}"
