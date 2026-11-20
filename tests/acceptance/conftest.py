import secrets
import shlex
import subprocess
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


def run_shampoo(workspace: Path, shampoo_binary: Path, command: str):
    args = command.split()
    assert args[0] == "shampoo"
    return subprocess.run(
        [str(shampoo_binary), *args[1:]],
        cwd=workspace,
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
        timeout=300,
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
def given_command_passed(workspace, shampoo_binary, command):
    result = run_shampoo(workspace, shampoo_binary, command)
    assert result.returncode == 0, (
        f"precondition failed: {command} exited {result.returncode}: "
        f"stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@given(parsers.parse('"{command}" has run'))
def given_command_ran(workspace, shampoo_binary, command):
    run_shampoo(workspace, shampoo_binary, command)


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
def when_running(workspace, shampoo_binary, command, run_result):
    result = run_shampoo(workspace, shampoo_binary, command)
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
    budget: str = "{ violations = 5, runs = 10 }",
    extra: str = "",
) -> None:
    rules = workspace / ".shampoo" / "rules"
    rules.mkdir(parents=True, exist_ok=True)
    (rules / f"{rule_id}.toml").write_text(
        f'rule = "{rule_id} holds."\n'
        f'check = "{check_id}"\n'
        f'transitions = ["{transition}"]\n'
        f"{extra}"
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


@given(parsers.parse('a rule guarding {transition:w} with check "{check_id}"'))
def given_rule(workspace, transition, check_id):
    write_rule(workspace, new_id(), transition, check_id)
