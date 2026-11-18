#!/bin/sh
set -u

usage='usage: judge-coverage.sh <base> [--update] <profile>...'
[ $# -ge 2 ] || { echo "$usage" >&2; exit 2; }
base=$1
shift
mode=''
if [ "$1" = "--update" ]; then
	mode=--update
	shift
fi
[ $# -ge 1 ] || { echo "$usage" >&2; exit 2; }
for profile in "$@"; do
	[ -f "$profile" ] || { echo "judge-coverage: no profile at $profile; run the tests first" >&2; exit 2; }
done
baseline=scripts/coverage-baseline.txt

head=$(awk '/^mode:/ { next } { stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END { for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }; print covered + 0, total + 0 }' "$@")
previous=$(git show "$base:$baseline" 2>/dev/null || echo "0 0")

set -- $head $previous
covered=$1 total=$2 base_covered=$3 base_total=$4

if [ $((covered * base_total)) -lt $((base_covered * total)) ]; then
	echo "coverage went down from $base_covered/$base_total to $covered/$total statements (.shampoo/rules/coverage.toml)" >&2
	exit 1
fi

if [ "$mode" = "--update" ]; then
	echo "$covered $total" > "$baseline"
fi
