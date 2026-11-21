#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/statements" ./statements || { echo "check-side-effect-statements: cannot build the check" >&2; exit 2; }
module=$(go list -m) || exit 2
files=$(scripts/list-files.sh '*.go' ':!:*_test.go' ':!:tools/**') || exit 2
printf '%s\n' "$files" | sed '/^$/d' | awk -v module="$module" 'FILENAME == ARGV[1] { dir = $0; sub("^" module "/?", "", dir); if (dir == "") dir = "."; side[dir] = 1; next }
	{ dir = $0; if (dir ~ /\//) sub(/\/[^\/]*$/, "", dir); else dir = "."; if (dir in side) print }' scripts/side-effect-packages.txt - | "$bin/statements"
