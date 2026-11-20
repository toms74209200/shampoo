import shlex

from pytest_bdd import given, parsers, scenarios

from conftest import new_id, write_check, write_rule

scenarios("../../features/exceptions.feature")


@given(parsers.parse('a check "{check_id}" that writes {count:d} exceptions to "{path}"'))
def given_check_writing_exceptions(workspace, check_id, count, path):
    quoted = shlex.quote(path)
    write_check(
        workspace,
        check_id,
        f"mkdir -p $(dirname {quoted}) && for i in $(seq {count}); do echo \"main.go:$i reason\"; done > {quoted}",
    )


@given(parsers.re(r'a rule guarding (?P<transition>\w+) with check "(?P<check_id>[^"]+)" and its exceptions in "(?P<path>[^"]+)"( limited to (?P<limit>\d+))?$'))
def given_rule_with_exceptions(workspace, transition, check_id, path, limit):
    declared = f'file = "{path}"' + (f", limit = {limit}" if limit else "")
    write_rule(workspace, new_id(), transition, check_id, extra=f"exceptions = {{ {declared} }}\n")
