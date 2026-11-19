#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

GOBIN=$bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 || { echo "golangci-lint: cannot install" >&2; exit 2; }
"$bin/golangci-lint" "$@"
