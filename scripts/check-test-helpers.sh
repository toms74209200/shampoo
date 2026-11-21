#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/testhelpers" ./testhelpers || { echo "check-test-helpers: cannot build the check" >&2; exit 2; }
files=$(scripts/list-files.sh '*_test.go') || exit 2
printf '%s\n' "$files" | sed '/^$/d' | "$bin/testhelpers"
