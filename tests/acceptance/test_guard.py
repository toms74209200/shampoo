import json
import secrets
import shlex
import subprocess
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

scenarios("../../features/guard.feature")

VIOLATION = '{ other = "task-failure" }'


@pytest.fixture
def runs_log(tmp_path):
    return tmp_path / "runs.log"


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
    budget: str = "0.5",
    paths: str = "",
) -> None:
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    (rules / f"{rule_id}.toml").write_text(
        f'rule = "{rule_id} holds."\n'
        f'check = "{check_id}"\n'
        f'transitions = ["{transition}"]\n'
        f"{paths}"
        f"judge = {judge}\n"
        f"budget = {budget}\n"
    )


def new_id() -> str:
    return "rule-" + secrets.token_hex(4)


def record(runs_log: Path, check_id: str) -> str:
    return f"echo {shlex.quote(check_id)} >> {shlex.quote(str(runs_log))}"


@given(parsers.parse('a check "{check_id}" that exits {code:d}'))
def given_check_exits(workspace, check_id, code):
    write_check(workspace, check_id, f"exit {code}")


@given(parsers.parse('a check "{check_id}" that prints "{text}" and exits {code:d}'))
def given_check_prints(workspace, check_id, text, code):
    write_check(workspace, check_id, f"echo {shlex.quote(text)}; exit {code}")


@given(parsers.parse('a check "{check_id}" that records that it ran'))
def given_check_records(workspace, runs_log, check_id):
    write_check(workspace, check_id, record(runs_log, check_id))


@given(parsers.parse('a check "{check_id}" that needs "{needed}" and records that it ran'))
def given_check_needs(workspace, runs_log, check_id, needed):
    write_check(workspace, check_id, record(runs_log, check_id), f'needs = ["{needed}"]\n')


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}"'))
def given_rule(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id)


@given(parsers.parse('a rule "{rule_id}" guarding {transition:w} with check "{check_id}"'))
def given_named_rule(workspace, rule_id, transition, check_id):
    write_rule(workspace, rule_id, transition, check_id)


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" and a zero budget'))
def given_rule_zero_budget(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id, budget="0")


@given(parsers.parse('a rule "{rule_id}" guarding {transition:w} with check "{check_id}" and a zero budget'))
def given_named_rule_zero_budget(workspace, rule_id, transition, check_id):
    write_rule(workspace, rule_id, transition, check_id, budget="0")


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" judging exit {code:d} as a workflow failure'))
def given_rule_workflow_failure(workspace, transition, check_id, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        check_id,
        judge=f'{{ exit = {{ {code} = "workflow-failure", other = "task-failure" }} }}',
    )


@given(parsers.parse('a rule guarding {transition:w} for paths "{paths}" with check "{check_id}"'))
def given_rule_paths(workspace, transition, paths, check_id):
    write_rule(workspace, new_id(), transition, check_id, paths=f'paths = ["{paths}"]\n')


@given(
    parsers.parse(
        'a rule "{rule_id}" guarding {transition:w} with check "{check_id}" whose judge prints "{text}" and exits {code:d}'
    )
)
def given_rule_judge_fails(workspace, rule_id, transition, check_id, text, code):
    command = f"echo {shlex.quote(text)}; exit {code}"
    write_rule(workspace, rule_id, transition, check_id, judge=f"{{ command = {command!r}, exit = {VIOLATION} }}")


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}" whose judge exits {code:d} charging its budget'))
def given_rule_judge_budget(workspace, transition, check_id, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        check_id,
        judge=f'{{ command = "exit {code}", exit = {{ {code} = "budget", other = "workflow-failure" }} }}',
    )


def write_budgets(workspace: Path, text: str) -> None:
    (workspace / ".shampoo").mkdir(parents=True, exist_ok=True)
    (workspace / ".shampoo" / "budgets.toml").write_text(text)


@given(parsers.parse("budget settings with a fast burn of {burn:d} times over the latest {count:d} runs"))
def given_fast_burn(workspace, burn, count):
    write_budgets(workspace, f"fast = {{ runs = {count}, burn = {burn} }}\n")


@given(parsers.parse("budget settings with a window of {window:d} runs and a slow burn of {burn:d} time over the window"))
def given_slow_burn(workspace, window, burn):
    write_budgets(workspace, f"window = {window}\nslow = {{ burn = {burn} }}\n")


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


@then(parsers.parse('the statistics have a line for task "{task}"'))
def then_statistics_line(workspace, task):
    common = subprocess.run(
        ["git", "-C", str(workspace), "rev-parse", "--path-format=absolute", "--git-common-dir"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()
    stats = Path(common) / "shampoo" / "stats.jsonl"
    lines = [json.loads(line) for line in stats.read_text().splitlines()] if stats.exists() else []
    assert any(line.get("task") == task for line in lines), f"expected a line for {task}, got {lines}"
