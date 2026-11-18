#!/bin/sh
set -u

[ $# -eq 2 ] && [ "$1" = "--rev" ] || { echo "usage: check-scenario-commit.sh --rev <rev>" >&2; exit 2; }
rev=$2

files=$(git diff-tree --no-commit-id --name-only -r --root "$rev") || { echo "check-scenario-commit: cannot list files of $rev" >&2; exit 2; }
production=$(printf '%s\n' "$files" | grep -E '\.go$' | grep -vE '_test\.go$|^tools/|^tests/')
features=$(printf '%s\n' "$files" | grep -E '\.feature$')
[ -n "$production" ] && [ -n "$features" ] || exit 0

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT
go build -C tools -o "$work/scenarios" ./scenarios || { echo "check-scenario-commit: cannot build tools/scenarios" >&2; exit 2; }

changed=''
for f in $features; do
	git show "$rev^:$f" > "$work/before.feature" 2>/dev/null || : > "$work/before.feature"
	git show "$rev:$f" > "$work/after.feature" 2>/dev/null || : > "$work/after.feature"
	before=$(echo "$work/before.feature" | "$work/scenarios" --sha | sed 's/.*@sha://' | sort)
	after=$(echo "$work/after.feature" | "$work/scenarios" --sha | sed 's/.*@sha://' | sort)
	[ "$before" = "$after" ] || changed="$changed $f"
done

if [ -n "$changed" ]; then
	echo "a commit that changes scenario content changes no production code (.shampoo/rules/acceptance-tests.toml):$changed" >&2
	exit 1
fi
