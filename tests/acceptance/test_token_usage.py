import json
import secrets
import subprocess
from pathlib import Path

from pytest_bdd import given, parsers, scenarios, then

from conftest import run_shampoo

scenarios("../../features/token-usage.feature")

USAGE = r"(?P<input>\d+) input, (?P<creation>\d+) cache creation, (?P<read>\d+) cache read and (?P<output>\d+) output tokens"


def transcript_path(workspace: Path) -> Path:
    return workspace.parent / "transcript.jsonl"


def message(input, creation, read, output) -> str:
    entry = {
        "type": "assistant",
        "message": {
            "id": f"msg_{secrets.token_hex(8)}",
            "role": "assistant",
            "usage": {
                "input_tokens": int(input),
                "cache_creation_input_tokens": int(creation),
                "cache_read_input_tokens": int(read),
                "output_tokens": int(output),
            },
        },
    }
    return json.dumps(entry) + "\n"


@given(parsers.re("a transcript with a message using " + USAGE))
def given_transcript(workspace, input, creation, read, output):
    transcript_path(workspace).write_text(message(input, creation, read, output))


@given(parsers.re("a transcript with a message written twice using " + USAGE))
def given_transcript_twice(workspace, input, creation, read, output):
    line = message(input, creation, read, output)
    transcript_path(workspace).write_text(line + line)


@given(parsers.re("the transcript adds a message using " + USAGE))
def given_transcript_adds(workspace, input, creation, read, output):
    with transcript_path(workspace).open("a") as f:
        f.write(message(input, creation, read, output))


@given(parsers.parse('"{command}" has passed with the transcript'))
def given_passed_with_transcript(workspace, shampoo_binary, fake_bin, command):
    hook_input = json.dumps({"hook_event_name": "Stop", "transcript_path": str(transcript_path(workspace))})
    result = run_shampoo(workspace, shampoo_binary, fake_bin, command, stdin=hook_input)
    assert result.returncode == 0 and json.loads(result.stdout)["code"] == 0, (
        f"precondition failed: {command} exited {result.returncode}: "
        f"stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


def statistics(workspace: Path, task: str) -> list:
    common = subprocess.run(
        ["git", "-C", str(workspace), "rev-parse", "--path-format=absolute", "--git-common-dir"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()
    stats = Path(common) / "shampoo" / "stats.jsonl"
    lines = [json.loads(line) for line in stats.read_text().splitlines()] if stats.exists() else []
    return [line for line in lines if line.get("task") == task]


@then(parsers.re('the statistics of task "(?P<task>[^"]+)" have ' + USAGE))
def then_statistics_tokens(workspace, task, input, creation, read, output):
    want = {"input": int(input), "cache-creation": int(creation), "cache-read": int(read), "output": int(output)}
    got = [line.get("tokens") for line in statistics(workspace, task)]
    assert got == [want], f"got tokens {got!r}"


@then(parsers.parse('the statistics of task "{task}" have {count:d} changed line'))
def then_statistics_changed(workspace, task, count):
    got = [line.get("changed") for line in statistics(workspace, task)]
    assert got == [count], f"got changed lines {got!r}"


@then(parsers.parse('the statistics of task "{task}" have no tokens'))
def then_statistics_no_tokens(workspace, task):
    got = [line.get("tokens") for line in statistics(workspace, task)]
    assert got == [None], f"got tokens {got!r}"
