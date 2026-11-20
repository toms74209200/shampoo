#!/bin/sh
set -u

[ $# -eq 1 ] || { echo "usage: coverage-summary.sh <small profile>" >&2; exit 2; }
[ -f "$1" ] || { echo "coverage-summary: no profile at $1; run make test-small" >&2; exit 2; }
module=$(go list -m) || exit 2

echo "## Coverage"
echo
echo "| Package | Kind | Statements | Small-test coverage |"
echo "|---|---|---:|---:|"
awk -v module="$module/" 'FILENAME == ARGV[1] { excluded[$0] = 1; next }
	/^mode:/ { next }
	{ stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END {
		for (b in stmts) {
			pkg = b
			sub(/\/[^\/]*$/, "", pkg)
			total[pkg] += stmts[b]
			if (b in hit) covered[pkg] += stmts[b]
		}
		for (pkg in total) {
			name = pkg
			sub("^" module, "", name)
			if (pkg in excluded) {
				printf "| %s | side effects | %d | |\n", name, total[pkg]
			} else {
				printf "| %s | pure | %d | %.1f%% (%d/%d) |\n", name, total[pkg], 100 * covered[pkg] / total[pkg], covered[pkg], total[pkg]
			}
		}
	}' scripts/side-effect-packages.txt "$1" | sort
awk 'FILENAME == ARGV[1] { excluded[$0] = 1; next }
	/^mode:/ { next }
	{ stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END {
		for (b in stmts) {
			pkg = b
			sub(/\/[^\/]*$/, "", pkg)
			all += stmts[b]
			if (pkg in excluded) continue
			pure += stmts[b]
			if (b in hit) covered += stmts[b]
		}
		printf "\nShare of pure code: %.1f%% (%d/%d statements)\n\n", 100 * pure / all, pure, all
		printf "Small-test coverage of pure code: %.1f%% (%d/%d statements), at least 95%% required\n", 100 * covered / pure, covered, pure
	}' scripts/side-effect-packages.txt "$1"
