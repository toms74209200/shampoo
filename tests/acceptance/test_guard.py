import secrets
import shlex
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

scenarios("../../features/guard.feature")


@pytest.fixture
def runs_log(tmp_path):
    return tmp_path / "runs.log"


def write_rule(workspace: Path, rule_id: str, transition: str, check: str, extra: str) -> None:
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    (rules / f"{rule_id}.toml").write_text(
        f'rule = "{rule_id} holds."\n'
        f"check = {check!r}\n"
        f'transitions = ["{transition}"]\n'
        f'size = "small"\n'
        f"{extra}"
    )


def new_id() -> str:
    return "rule-" + secrets.token_hex(4)


@given(parsers.parse("a rule guarding {transition:w} whose check exits {code:d}"))
def given_rule_exits(workspace, transition, code):
    write_rule(workspace, new_id(), transition, f"exit {code}", 'exit = { other = "task-failure" }\n')


@given(
    parsers.parse(
        'a rule "{rule_id}" guarding {transition:w} whose check prints "{text}" and exits {code:d} as a task failure'
    )
)
def given_rule_prints(workspace, rule_id, transition, text, code):
    write_rule(
        workspace,
        rule_id,
        transition,
        f"echo {shlex.quote(text)}; exit {code}",
        f'exit = {{ {code} = "task-failure", other = "workflow-failure" }}\n',
    )


@given(parsers.parse("a rule guarding {transition:w} whose check exits {code:d} as a task failure"))
def given_rule_task_failure(workspace, transition, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        f"exit {code}",
        f'exit = {{ {code} = "task-failure", other = "workflow-failure" }}\n',
    )


@given(parsers.parse("a rule guarding {transition:w} whose check exits {code:d} as a workflow failure"))
def given_rule_workflow_failure(workspace, transition, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        f"exit {code}",
        f'exit = {{ {code} = "workflow-failure", other = "task-failure" }}\n',
    )


@given(parsers.parse("a rule guarding {transition:w} whose check exits {code:d} charging its budget"))
def given_rule_budget(workspace, transition, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        f"exit {code}",
        f'exit = {{ {code} = "budget", other = "workflow-failure" }}\nbudget = 0.1\n',
    )


@given(
    parsers.parse(
        'a rule guarding {transition:w} for paths "{paths}" whose check exits {code:d} as a task failure'
    )
)
def given_rule_paths(workspace, transition, paths, code):
    write_rule(
        workspace,
        new_id(),
        transition,
        f"exit {code}",
        f'paths = ["{paths}"]\nexit = {{ {code} = "task-failure", other = "workflow-failure" }}\n',
    )


@given(parsers.parse("a rule guarding {transition:w} whose check records that it ran"))
def given_rule_records(workspace, runs_log, transition):
    write_rule(
        workspace,
        new_id(),
        transition,
        f"echo ran >> {shlex.quote(str(runs_log))}",
        'exit = { other = "task-failure" }\n',
    )


@then(parsers.parse("the check has run {times}"))
def then_check_ran(runs_log, times):
    expected = {"once": 1, "twice": 2}[times]
    runs = runs_log.read_text().splitlines() if runs_log.exists() else []
    assert len(runs) == expected, f"expected {expected} runs, got {len(runs)}"
