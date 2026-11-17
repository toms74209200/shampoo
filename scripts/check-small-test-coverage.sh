#!/bin/sh
set -u
threshold=95

packages=$(go list ./...) || { echo "check-small-test-coverage: go list failed" >&2; exit 2; }
measured=$(printf '%s\n' "$packages" | grep -vxF -f scripts/side-effect-packages.txt | paste -sd, -)
profile=$(mktemp) || exit 2
trap 'rm -f "$profile"' EXIT

if [ -n "$measured" ]; then
	go test -tags=small -coverpkg="$measured" -coverprofile="$profile" ./... >/dev/null || { echo "small tests failed, so coverage cannot be measured (.shampoo/rules/small-test-coverage.toml)" >&2; exit 1; }
fi

result=$(awk 'NR > 1 { stmts[$1] = $2; if ($3 > 0) hit[$1] = 1 }
	END { for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }; print covered + 0, total + 0 }' "$profile")
covered=${result% *}
total=${result#* }

if [ "$total" -gt 0 ] && [ $((covered * 100)) -lt $((threshold * total)) ]; then
	echo "small-test coverage of code outside side-effect packages is $covered/$total statements, below ${threshold}% (.shampoo/rules/small-test-coverage.toml)" >&2
	exit 1
fi
