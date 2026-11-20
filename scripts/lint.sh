#!/bin/sh
set -u

root=$(git rev-parse --show-toplevel) || exit 2
mkdir -p "$root/out/exceptions" || exit 2
git -C "$root" grep -n -E '//nolint:[a-z]' -- '*.go' > "$root/out/exceptions/lint.txt"
[ $? -le 1 ] || exit 2
"$root/scripts/golangci-lint.sh" run --build-tags small,medium,large || exit
cd "$root/tools" && "$root/scripts/golangci-lint.sh" run --build-tags small,medium,large
