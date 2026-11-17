#!/bin/sh
set -u

[ $# -ge 1 ] || { echo "usage: check-coverage.sh <base> [--update]" >&2; exit 2; }
base=$1
baseline=scripts/coverage-baseline.txt

profile=$(mktemp) || exit 2
trap 'rm -f "$profile"' EXIT

packages=$(go list ./... | paste -sd, -) || { echo "check-coverage: go list failed" >&2; exit 2; }
go test -tags=small -coverpkg="$packages" -coverprofile="$profile" ./... >/dev/null || { echo "tests failed, so coverage cannot be measured (.shampoo/rules/coverage.toml)" >&2; exit 1; }

head=$(awk 'NR > 1 { stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END { for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }; print covered + 0, total + 0 }' "$profile")
previous=$(git show "$base:$baseline" 2>/dev/null || echo "0 0")

set -- $head $previous "${2:-}"
covered=$1 total=$2 base_covered=$3 base_total=$4 mode=$5

if [ $((covered * base_total)) -lt $((base_covered * total)) ]; then
	echo "coverage went down from $base_covered/$base_total to $covered/$total statements (.shampoo/rules/coverage.toml)" >&2
	exit 1
fi

if [ "$mode" = "--update" ]; then
	echo "$covered $total" > "$baseline"
fi
