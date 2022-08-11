#!/bin/sh

set -eu

PROJECTS="$*"
COMMANDS="tidy get build test"

expand() {
	local prefix="$1" x=
	shift
	for x; do
		echo "$prefix-$x"
	done | tr '\n' ' '
}

for cmd in $COMMANDS; do
	all="$(expand $cmd root $PROJECTS)"
	cat <<EOT
.PHONY: $cmd $all
$cmd: $all

EOT
	case "$cmd" in
	tidy)	call="\$(GO) mod tidy || true" ;;
	*)      call="\$(GO) $cmd -v ./..." ;;
	esac

	for x in . $PROJECTS; do
		if [ "$x" = . ]; then
			k="root"
			cd=
		else
			k="$x"
			cd="cd '$x' && "
		fi
		deps="$(sed -n -e 's|^.*=> \.\?\./\([^/]\+\).*$|\1|p' "$x/go.mod" | tr '\n' ' ')"

		cat <<EOT
$cmd-$k:${deps:+ $(expand $cmd $deps)}
	$cd$call

EOT
	done
done
