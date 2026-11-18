#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/fixedvalues" ./fixedvalues || { echo "check-fixed-values: cannot build the check" >&2; exit 2; }
files=$(git ls-files '*_test.go') || { echo "check-fixed-values: git ls-files failed" >&2; exit 2; }
printf '%s\n' "$files" | sed '/^$/d' | "$bin/fixedvalues"
