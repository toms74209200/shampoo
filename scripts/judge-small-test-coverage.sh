#!/bin/sh
set -u
threshold=95

[ $# -eq 1 ] || { echo "usage: judge-small-test-coverage.sh <profile>" >&2; exit 2; }
[ -f "$1" ] || { echo "judge-small-test-coverage: no profile at $1; run make test-small" >&2; exit 2; }

result=$(awk 'FILENAME == ARGV[1] { excluded[$0] = 1; next }
	/^mode:/ { next }
	{
		pkg = $1
		sub(/\/[^\/]*$/, "", pkg)
		if (pkg in excluded) next
		stmts[$1] = $2
		if ($3 > 0) hit[$1] = 1
	}
	END { for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }; print covered + 0, total + 0 }' scripts/side-effect-packages.txt "$1")
covered=${result% *}
total=${result#* }

if [ "$total" -gt 0 ] && [ $((covered * 100)) -lt $((threshold * total)) ]; then
	echo "small-test coverage of code outside side-effect packages is $covered/$total statements, below ${threshold}% (.shampoo/rules/small-test-coverage.toml)" >&2
	exit 1
fi
