#!/bin/sh
set -u

binaries=$(mktemp -d) || exit 2
trap 'rm -rf "$binaries"' EXIT

packages=$(go list ./...) || { echo "check-side-effects: go list failed" >&2; exit 2; }
for p in $packages; do
	CGO_ENABLED=0 go test -c -tags=small -o "$binaries/$(printf '%s' "$p" | tr '/' '_').test" "$p" >/dev/null || { echo "check-side-effects: cannot build small tests of $p" >&2; exit 1; }
done

failed=''
for b in "$binaries"/*.test; do
	[ -e "$b" ] || continue
	bwrap --unshare-all --die-with-parent --clearenv --ro-bind "$b" /test --proc /proc --dev /dev --chdir / /test -test.timeout=60s >/dev/null 2>&1 || failed="$failed $(basename "$b" .test)"
done

if [ -n "$failed" ]; then
	echo "small tests failed in isolation, without network, file system or environment (.shampoo/rules/side-effects.toml):$failed" >&2
	exit 1
fi
