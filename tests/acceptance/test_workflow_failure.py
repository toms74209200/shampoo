from pytest_bdd import given, parsers, scenarios

from conftest import write_settings

scenarios("../../features/workflow-failure.feature")


@given(parsers.re(r"settings stopping workflow failures at (?P<violations>\d+) in the latest (?P<runs>\d+) runs?"))
def given_workflow_failure_stop(workspace, settings_sections, violations, runs):
    settings_sections.setdefault("workflow-failure", {})["stop"] = {"violations": int(violations), "runs": int(runs)}
    write_settings(workspace, settings_sections)
