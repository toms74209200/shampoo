import json
import shlex
import subprocess
from pathlib import Path

import pytest
from pytest_bdd import given, parsers, scenarios, then

scenarios("../../features/run.feature")


@pytest.fixture
def agent_log(tmp_path):
    log = tmp_path / "agent"
    log.mkdir()
    return log


@pytest.fixture
def settings(workspace):
    return {}


def write_settings(workspace: Path, settings: dict) -> None:
    (workspace / ".shampoo").mkdir(parents=True, exist_ok=True)
    lines = ["[orchestrator]"] + [f"{key} = {json.dumps(value)}" for key, value in settings.items()]
    (workspace / ".shampoo" / "settings.toml").write_text("\n".join(lines) + "\n")


@given(parsers.parse('a task "{task_id}" with the definition of done "{dod}"'))
def given_task(workspace, settings, task_id, dod):
    task = json.dumps({"id": task_id, "dod": dod})
    settings["tasks"] = f"printf '%s\\n' {shlex.quote(task)}"
    write_settings(workspace, settings)


@given("an agent that records where it works")
def given_recording_agent(workspace, settings, agent_log):
    log = shlex.quote(str(agent_log))
    script = (
        f'printf "%s" "$1" > {log}/dod; '
        f"git rev-parse --abbrev-ref HEAD > {log}/branch; "
        f"git rev-parse --show-toplevel > {log}/toplevel"
    )
    settings["agent"] = f"sh -c {shlex.quote(script)} agent"
    write_settings(workspace, settings)


@given(parsers.parse('an agent that changes "{path}"'))
def given_changing_agent(workspace, settings, path):
    settings["agent"] = f"sh -c {shlex.quote(f'echo changed >> {shlex.quote(path)}')} agent"
    write_settings(workspace, settings)


@then(parsers.parse('the agent received "{dod}"'))
def then_agent_received(agent_log, dod):
    received = (agent_log / "dod").read_text() if (agent_log / "dod").exists() else None
    assert received == dod, f"expected the agent to receive {dod!r}, got {received!r}"


@then(parsers.parse('the agent worked on the branch "{branch}"'))
def then_agent_branch(agent_log, branch):
    got = (agent_log / "branch").read_text().strip() if (agent_log / "branch").exists() else None
    assert got == branch, f"expected branch {branch!r}, got {got!r}"


@then("the agent worked outside the repository's own working tree")
def then_agent_outside(workspace, agent_log):
    got = (agent_log / "toplevel").read_text().strip() if (agent_log / "toplevel").exists() else None
    assert got is not None and Path(got).resolve() != workspace.resolve(), f"expected a worktree of its own, got {got!r}"
