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

set -- $(awk -v c="$covered" -v t="$total" -v bc="$base_covered" -v bt="$base_total" 'BEGIN {
	p = (bc + 1) / (bt + 2)
	q = t > 0 ? c / t : 1
	bound = t > 0 ? p - sqrt(p * (1 - p) / t) : 0
	printf "%.1f %.1f %d\n", q * 100, bound * 100, q < bound
}')
share=$1 bound=$2 dropped=$3

report="share of pure code: $covered/$total statements ($share%), base $base_covered/$base_total, one standard deviation below the base at $bound%"
echo "$report"
[ -n "${GITHUB_STEP_SUMMARY:-}" ] && echo "$report" >> "$GITHUB_STEP_SUMMARY"

if [ "$mode" = "--update" ]; then
	echo "$covered $total" > "$baseline"
fi

if [ "$dropped" -eq 1 ]; then
	echo "the share of pure code dropped more than one standard deviation, which charges the structure budget (.shampoo/rules/coverage.toml)" >&2
	exit 3
fi
