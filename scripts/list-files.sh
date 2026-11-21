#!/bin/sh
set -u

files=$(git ls-files --cached --others --exclude-standard "$@") || { echo "list-files: git ls-files failed" >&2; exit 2; }
printf '%s\n' "$files" | while IFS= read -r file; do
	if [ -f "$file" ]; then
		printf '%s\n' "$file"
	fi
done
