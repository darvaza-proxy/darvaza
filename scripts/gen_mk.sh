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
	tidy)
		call="$(cat <<EOT | sed -e '/^$/d;'
\$(GO) vet ./...
\$(GO) mod tidy
EOT
)"
		;;
	*)      call="\$(GO) $cmd -v ./..." ;;
	esac

	# tidy up call

	case "$cmd" in
	build|test)
		sequential=true ;;
	*)
		sequential=false ;;
	esac

	for x in . $PROJECTS; do
		if [ "$x" = . ]; then
			k="root"
			cd=
		else
			k="$x"
			cd="cd '$x' \&\& "
		fi

		if [ "$k" = root -a "$cmd" = build ]; then
			callx="\$(GO) $cmd -o \$(TMPDIR)/ -v ./..."
		else
			callx="$call"
		fi

		if $sequential; then
			deps="$(sed -n -e 's|^.*=> \.\?\./\([^/]\+\).*$|\1|p' "$x/go.mod" | tr '\n' ' ')"
		else
			deps=
		fi

		cat <<EOT
$cmd-$k:${deps:+ $(expand $cmd $deps)}
$(echo "$callx" | sed -e "/^$/d;" -e "s|^|\t$cd|")

EOT
	done
done
