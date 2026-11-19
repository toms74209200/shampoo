#!/bin/sh
set -u

root=$(git rev-parse --show-toplevel) || exit 2
"$root/scripts/golangci-lint.sh" run --build-tags small,medium,large || exit
cd "$root/tools" && "$root/scripts/golangci-lint.sh" run --build-tags small,medium,large
