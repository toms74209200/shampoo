#!/bin/sh
set -u

files=$(scripts/list-files.sh '*.go') || exit 2

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
