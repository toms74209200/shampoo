import os
import subprocess

from pytest_bdd import given, parsers, scenarios, then, when

from conftest import REPO_ROOT

scenarios("../../features/help.feature")


@given(parsers.parse('shampoo built as version "{version}"'), target_fixture="versioned_binary")
def given_versioned_binary(tmp_path, version):
    binary = tmp_path / "shampoo-versioned"
    subprocess.run(
        ["go", "build", "-ldflags", f"-X github.com/toms74209200/shampoo/internal/cli.Version={version}", "-o", str(binary), "."],
        cwd=REPO_ROOT,
        check=True,
    )
    return binary


@when(parsers.parse('I run that shampoo with "{argument}"'))
def when_running_versioned(workspace, versioned_binary, run_result, argument):
    result = subprocess.run([str(versioned_binary), argument], cwd=workspace, capture_output=True, text=True, timeout=60, env=os.environ)
    run_result.update({"stdout": result.stdout, "stderr": result.stderr, "returncode": result.returncode})


@then(parsers.parse('stdout is "{text}"'))
def then_stdout_is(run_result, text):
    assert run_result["stdout"] == text + "\n", f"expected {text!r} on stdout, got {run_result['stdout']!r}"
