#!/bin/sh
set -u

modules=$(go list -deps -test -tags small,medium,large -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./... | sort -u) || { echo "check-dependencies: go list failed" >&2; exit 2; }
allowed=$(cut -d' ' -f1 scripts/allowed-modules.txt) || { echo "check-dependencies: cannot read scripts/allowed-modules.txt" >&2; exit 2; }

missing=$(printf '%s\n' "$modules" | sed '/^$/d' | while IFS= read -r m; do
	printf '%s\n' "$allowed" | grep -qxF "$m" || printf '%s\n' "$m"
done)

if [ -n "$missing" ]; then
	echo "every module needs an entry in scripts/allowed-modules.txt, with its reason (.shampoo/rules/dependencies.toml):" >&2
	printf '%s\n' "$missing" >&2
	exit 1
fi
