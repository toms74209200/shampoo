#!/bin/sh
set -u

root=$(git rev-parse --show-toplevel) || exit 1
cd "$root" || exit 1
mkdir -p out || exit 1
if go build -o out/shampoo.next . 2>out/shampoo.build.log; then
	mv out/shampoo.next out/shampoo
elif [ ! -x out/shampoo ]; then
	echo "hook: cannot build shampoo and no earlier build exists:" >&2
	cat out/shampoo.build.log >&2
	exit 1
fi
case "${1:-}" in
guard)
	out/shampoo guard "${2:-edit}"
	case $? in
	0) exit 0 ;;
	1 | 3) exit 2 ;;
	*) exit 1 ;;
	esac
	;;
policies)
	out/shampoo policies || exit 1
	;;
*)
	echo "usage: hook.sh guard <transition> | hook.sh policies" >&2
	exit 1
	;;
esac
