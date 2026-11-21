#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/testhelpers" ./testhelpers || { echo "check-test-helpers: cannot build the check" >&2; exit 2; }
files=$(git ls-files --cached --others --exclude-standard '*_test.go') || { echo "check-test-helpers: git ls-files failed" >&2; exit 2; }
files=$(printf '%s\n' "$files" | while IFS= read -r file; do if [ -f "$file" ]; then printf '%s\n' "$file"; fi; done)
printf '%s\n' "$files" | sed '/^$/d' | "$bin/testhelpers"
