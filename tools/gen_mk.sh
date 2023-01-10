#!/bin/sh

set -eu

PROJECTS="$*"
COMMANDS="tidy get build test up"

expand() {
	local prefix="$1" x=
	shift
	for x; do
		echo "$prefix-$x"
	done | tr '\n' ' '
}

for cmd in $COMMANDS; do
	all="$(expand $cmd root $PROJECTS)"
	depsx=

	cat <<EOT
.PHONY: $cmd $all
$cmd: $all

EOT

	case "$cmd" in
	tidy)
		call="$(cat <<EOT | sed -e '/^$/d;'
\$(GO) mod tidy
\$(GO) vet ./...
\$(REVIVE) \$(REVIVE_RUN_ARGS) ./...
EOT
)"
		depsx="\$(REVIVE)"
		;;
	up)
		call="\$(GO) get -u -v ./...
\$(GO) mod tidy"
		;;
	*)
		call="\$(GO) $cmd -v ./..."
		;;
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

		callx="$call"

		if [ "$k" = root ]; then
			# special case

			case "$cmd" in
			build)
				cmdx="$cmd -o \$(TMPDIR)/"
				;;
			get|up)
				cmdx="get -tags tools"
				;;
			*)
				cmdx=
				;;
			esac

			[ -z "$cmdx" ] || cmdx="\$(GO) $cmdx -v ./..."

			if [ "up" = "$cmd" ]; then
				callx="$cmdx
\$(GO) mod tidy
\$(GO) install -v \$(REVIVE_INSTALL_URL)"
			elif [ "tidy" = "$cmd" ]; then
				exclude=
				for x in $PROJECTS; do
					exclude="${exclude:+$exclude }-exclude ./$x/..."
				done
				callx=$(echo "$callx" | sed -e "s;\(REVIVE)\);\1 $exclude;")
			elif [ -n "$cmdx" ]; then
				callx="$cmdx"
			fi
		fi

		if $sequential; then
			deps="$(sed -n -e 's|^.*=> \.\?\./\([^/]\+\).*$|\1|p' "$x/go.mod" | tr '\n' ' ')"
		else
			deps=
		fi

		cat <<EOT
$cmd-$k:${depsx:+ $depsx}${deps:+ $(expand $cmd $deps)}
$(echo "$callx" | sed -e "/^$/d;" -e "s|^|\t@$cd|")

EOT
	done
done
