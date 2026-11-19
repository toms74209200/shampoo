#!/bin/sh
set -u

root=$(git rev-parse --show-toplevel) || exit 2
cd "$root" || exit 2
strace -qq -o /dev/null true || { echo "test-small: strace cannot trace processes here" >&2; exit 2; }
rm -rf out/traces && mkdir -p out/traces || exit 2
TZ= CGO_ENABLED=0 TRACE_DIR=$root/out/traces go test -tags=small -exec "$root/scripts/trace-exec.sh" -coverpkg=./... -coverprofile=out/small.cover ./...
