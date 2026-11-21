#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/scenarios" ./scenarios || { echo "check-scenarios: cannot build the check" >&2; exit 2; }
files=$(scripts/list-files.sh '*.feature') || exit 2
printf '%s\n' "$files" | sed '/^$/d' | "$bin/scenarios" "$@"
