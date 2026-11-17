#!/bin/sh
set -u

[ $# -ge 2 ] || { echo "usage: judge-coverage.sh <profile> <base> [--update]" >&2; exit 2; }
[ -f "$1" ] || { echo "judge-coverage: no profile at $1; run make test-small" >&2; exit 2; }
baseline=scripts/coverage-baseline.txt

head=$(awk '/^mode:/ { next } { stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END { for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }; print covered + 0, total + 0 }' "$1")
previous=$(git show "$2:$baseline" 2>/dev/null || echo "0 0")

set -- $head $previous "${3:-}"
covered=$1 total=$2 base_covered=$3 base_total=$4 mode=$5

if [ $((covered * base_total)) -lt $((base_covered * total)) ]; then
	echo "coverage went down from $base_covered/$base_total to $covered/$total statements (.shampoo/rules/coverage.toml)" >&2
	exit 1
fi

if [ "$mode" = "--update" ]; then
	echo "$covered $total" > "$baseline"
fi
