#!/bin/sh
set -u

files=$(git ls-files --cached --others --exclude-standard '*.go') || { echo "check-exceptions: git ls-files failed" >&2; exit 2; }
files=$(printf '%s\n' "$files" | while IFS= read -r file; do if [ -f "$file" ]; then printf '%s\n' "$file"; fi; done)

found=''
if [ -n "$files" ]; then
	found=$(printf '%s\n' "$files" | xargs grep -nE '//[[:space:]]*lint:(file-)?ignore|#nosec' ; true)
fi
excluded=$(grep -nE '^[[:space:]]*exclusions:' .golangci.yml | sed 's/^/.golangci.yml:/')
found=$(printf '%s\n%s' "$found" "$excluded" | sed '/^$/d')

if [ -n "$found" ]; then
	echo "exceptions other than //nolint:<linter> // <reason> are prohibited (.shampoo/rules/exceptions.toml):" >&2
	printf '%s\n' "$found" >&2
	exit 1
fi
