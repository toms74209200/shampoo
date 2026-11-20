from pytest_bdd import given, parsers, scenarios

from conftest import write_settings

scenarios("../../features/next.feature")


@given(parsers.parse('a task source "{path}" that does not exist'))
def given_missing_task_source(workspace, settings_sections, path):
    settings_sections.setdefault("orchestrator", {})["tasks"] = path
    write_settings(workspace, settings_sections)
