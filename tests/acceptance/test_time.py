from pytest_bdd import given, parsers, scenarios

from conftest import write_check, write_settings

scenarios("../../features/time.feature")


@given(parsers.parse('a check "{check_id}" that takes {seconds} seconds and exits {code:d}'))
def given_slow_check(workspace, check_id, seconds, code):
    write_check(workspace, check_id, f"sleep {float(seconds)}; exit {code}")


@given(parsers.re(r"settings limiting small checks to (?P<seconds>\d+) seconds?"))
def given_small_limit(workspace, settings_sections, seconds):
    settings_sections.setdefault("test-sizes.small", {})["max-seconds"] = int(seconds)
    write_settings(workspace, settings_sections)
