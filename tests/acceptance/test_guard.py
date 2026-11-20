import json
import secrets
import shlex
import subprocess
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

from conftest import VIOLATION, new_id, record, write_check, write_rule

scenarios("../../features/guard.feature")

@pytest.fixture
def runs_log(tmp_path):
    return tmp_path / "runs.log"


@given(parsers.parse('a check "{check_id}" that prints "{text}" and exits {code:d}'))
def given_check_prints(workspace, check_id, text, code):
    write_check(workspace, check_id, f"echo {shlex.quote(text)}; exit {code}")


@given(parsers.parse('a check "{check_id}" that records that it ran'))
def given_check_records(workspace, runs_log, check_id):
    write_check(workspace, check_id, record(runs_log, check_id))


@given(parsers.parse('a check "{check_id}" that records that it ran and exits {code:d}'))
def given_check_records_exits(workspace, runs_log, check_id, code):
    write_check(workspace, check_id, f"{record(runs_log, check_id)}; exit {code}")


@given(parsers.parse('a check "{check_id}" that needs "{needed}" and records that it ran'))
def given_check_needs(workspace, runs_log, check_id, needed):
    write_check(workspace, check_id, record(runs_log, check_id), f'needs = ["{needed}"]\n')


@given(parsers.parse('a rule "{rule_id}" guarding {transition:w} with check "{check_id}"'))
def given_named_rule(workspace, rule_id, transition, check_id):
    write_rule(workspace, rule_id, transition, check_id)


@given(parsers.parse('a rule "{rule_id}" guarding {transition:w} with check "{check_id}" and a zero budget'))
def given_named_rule_zero_budget(workspace, rule_id, transition, check_id):
    write_rule(workspace, rule_id, transition, check_id, budget="{ violations = 0, runs = 1 }")


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
    write_rule(workspace, new_id(), transition, check_id, extra=f'paths = ["{paths}"]\n')


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


@given(
    parsers.parse(
        'a rule "{rule_id}" guarding {transition:w} with check "{check_id}" and a stop at {violations:d} violations in the latest {count:d} runs'
    )
)
def given_rule_stop(workspace, rule_id, transition, check_id, violations, count):
    write_rule(workspace, rule_id, transition, check_id, extra=f"stop = {{ violations = {violations}, runs = {count} }}\n")


@given(
    parsers.parse(
        'a rule "{rule_id}" guarding {transition:w} with check "{check_id}" and a budget of {violations:d} violation in {count:d} runs'
    )
)
def given_rule_budget_counts(workspace, rule_id, transition, check_id, violations, count):
    write_rule(workspace, rule_id, transition, check_id, budget=f"{{ violations = {violations}, runs = {count} }}")


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
