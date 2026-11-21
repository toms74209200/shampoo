import json

from pytest_bdd import given, parsers, scenarios, then, when

from conftest import next_action, run_shampoo, write_settings

scenarios("../../features/review.feature")


def findings_input(findings: list) -> str:
    return json.dumps({"findings": findings})


@given(parsers.parse('settings requiring "{key}" "{value}" of the review'))
def given_review_requirement(workspace, settings_sections, key, value):
    settings_sections.setdefault("state-machine.review.requirements", {})[key] = value
    write_settings(workspace, settings_sections)


def record_review(workspace, shampoo_binary, fake_bin, findings: list) -> None:
    result = run_shampoo(workspace, shampoo_binary, fake_bin, "shampoo review", stdin=findings_input(findings))
    assert result.returncode == 0, (
        f"precondition failed: shampoo review exited {result.returncode}: "
        f"stdout={result.stdout!r}, stderr={result.stderr!r}"
    )


@given(parsers.parse('the review found "{comment}" on policy "{policy_id}"'))
def given_review_found(workspace, shampoo_binary, fake_bin, comment, policy_id):
    record_review(workspace, shampoo_binary, fake_bin, [{"policy": policy_id, "comment": comment}])


@given(parsers.re(r'the review found (?P<count>\d+) findings? on policy "(?P<policy_id>[^"]+)"'))
def given_review_found_count(workspace, shampoo_binary, fake_bin, count, policy_id):
    findings = [{"policy": policy_id, "comment": f"finding {i}"} for i in range(int(count))]
    record_review(workspace, shampoo_binary, fake_bin, findings)


@given(parsers.parse("settings allowing {count:d} reviews whose findings do not drop"))
def given_review_attempts(workspace, settings_sections, count):
    settings_sections.setdefault("state-machine.review", {})["max-failed-attempts"] = count
    write_settings(workspace, settings_sections)


@given("the review found nothing")
def given_review_found_nothing(workspace, shampoo_binary, fake_bin):
    record_review(workspace, shampoo_binary, fake_bin, [])


@when(parsers.parse('I run "shampoo review" with a finding on policy "{policy_id}"'))
def when_reviewing(workspace, shampoo_binary, fake_bin, run_result, policy_id):
    result = run_shampoo(
        workspace, shampoo_binary, fake_bin, "shampoo review",
        stdin=findings_input([{"policy": policy_id, "comment": "found"}]),
    )
    run_result.update({"stdout": result.stdout, "stderr": result.stderr, "returncode": result.returncode})


@then(parsers.parse('the next action has the policies "{policy_ids}"'))
def then_policies(run_result, policy_ids):
    got = next_action(run_result).get("policies")
    assert got == policy_ids.split(", "), f"got policies {got!r}"


@then(parsers.parse('the next action has the requirement "{key}" "{value}"'))
def then_requirement(run_result, key, value):
    got = next_action(run_result).get("requirements", {}).get(key)
    assert got == value, f"got requirement {key} {got!r}"


@then(parsers.parse('the next action has the finding "{comment}" on policy "{policy_id}"'))
def then_finding(run_result, comment, policy_id):
    got = next_action(run_result).get("findings")
    assert got == [{"policy": policy_id, "comment": comment}], f"got findings {got!r}"


@then("the next action has no findings")
def then_no_findings(run_result):
    got = next_action(run_result).get("findings")
    assert not got, f"got findings {got!r}"
