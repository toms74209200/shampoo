#!/bin/sh
set -u

[ $# -eq 1 ] || { echo "usage: judge-side-effects.sh <traces>" >&2; exit 2; }
[ -d "$1" ] || { echo "judge-side-effects: no traces in $1; run make test-small" >&2; exit 2; }

found=$(for trace in "$1"/*.trace; do
	[ -e "$trace" ] || continue
	awk -v test="$(basename "$trace" | sed 's/\.[0-9]*\.trace$//')" '
	function allowed(path) {
		return index(path, root) == 1 || index(path, "/proc/self/") == 1 || index(path, "/sys/fs/cgroup/") == 1 || index(path, "/sys/kernel/mm/transparent_hugepage/") == 1
	}
	{
		line = $0
		sub(/^[0-9]+ +/, "", line)
		call = line
		sub(/\(.*/, "", call)
	}
	call == "execve" && root == "" {
		match(line, /"[^"]*"/)
		root = substr(line, RSTART + 1, RLENGTH - 2)
		sub(/\/[^\/]+\/[^\/]+$/, "/", root)
		next
	}
	call ~ /^(socket|socketpair|connect|bind|listen|accept|accept4|sendto|recvfrom|sendmsg|recvmsg|sendmmsg|recvmmsg|shutdown|getsockopt|setsockopt|getsockname|getpeername)$/ {
		print test ": " line
		next
	}
	{
		rest = line
		while (match(rest, /"[^"]*"/)) {
			path = substr(rest, RSTART + 1, RLENGTH - 2)
			rest = substr(rest, RSTART + RLENGTH)
			if (!allowed(path)) {
				print test ": " line
				break
			}
		}
	}' "$trace"
done)

if [ -n "$found" ]; then
	echo "small tests touched the file system or the network (.shampoo/rules/side-effects.toml):" >&2
	printf '%s\n' "$found" >&2
	exit 1
fi
