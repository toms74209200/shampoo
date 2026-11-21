#!/bin/sh
set -u

bin=$(mktemp -d) || exit 2
trap 'rm -rf "$bin"' EXIT

go build -C tools -o "$bin/whitebox" ./whitebox || { echo "check-white-box: cannot build the check" >&2; exit 2; }
files=$(scripts/list-files.sh '*.go' ':!:tools/**') || exit 2
printf '%s\n' "$files" | sed '/^$/d' | "$bin/whitebox"
