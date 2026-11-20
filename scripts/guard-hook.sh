#!/bin/sh
set -u

root=$(git rev-parse --show-toplevel) || exit 1
cd "$root" || exit 1
mkdir -p out || exit 1
if go build -o out/shampoo.next . 2>out/shampoo.build.log; then
	mv out/shampoo.next out/shampoo
elif [ ! -x out/shampoo ]; then
	echo "guard-hook: cannot build shampoo and no earlier build exists:" >&2
	cat out/shampoo.build.log >&2
	exit 1
fi
out/shampoo guard "${1:-edit}"
case $? in
0) exit 0 ;;
1 | 3) exit 2 ;;
*) exit 1 ;;
esac
