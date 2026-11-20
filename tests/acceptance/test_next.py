from pytest_bdd import given, parsers, scenarios

from conftest import write_settings

scenarios("../../features/next.feature")


@given(parsers.parse('a task source "{path}" that does not exist'))
def given_missing_task_source(workspace, settings_sections, path):
    settings_sections.setdefault("orchestrator", {})["tasks"] = path
    write_settings(workspace, settings_sections)


@given("no pull request is open")
def given_no_pull_request(fake_bin):
    gh = fake_bin / "gh"
    gh.write_text("#!/bin/sh\necho 'no pull requests found' >&2\nexit 1\n")
    gh.chmod(0o755)
