module darvaza.org/darvaza/server

go 1.21

require (
	darvaza.org/core v0.15.6
	darvaza.org/darvaza/shared v0.6.2
	darvaza.org/slog v0.5.15
	darvaza.org/slog/handlers/cblog v0.5.13 // indirect
)

require (
	github.com/miekg/dns v1.1.62
	github.com/naoina/toml v0.1.1
)

require (
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)

replace darvaza.org/darvaza/shared => ../shared
