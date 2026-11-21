#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/scenarios" ./scenarios || { echo "check-scenarios: cannot build the check" >&2; exit 2; }
files=$(git ls-files --cached --others --exclude-standard '*.feature') || { echo "check-scenarios: git ls-files failed" >&2; exit 2; }
files=$(printf '%s\n' "$files" | while IFS= read -r file; do if [ -f "$file" ]; then printf '%s\n' "$file"; fi; done)
printf '%s\n' "$files" | sed '/^$/d' | "$bin/scenarios" "$@"
