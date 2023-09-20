module darvaza.org/darvaza/server

go 1.19

replace darvaza.org/darvaza/shared => ../shared

require (
	darvaza.org/darvaza/shared v0.5.9
	darvaza.org/slog v0.5.4
	github.com/miekg/dns v1.1.55
	github.com/naoina/toml v0.1.1
)

require (
	darvaza.org/core v0.9.9 // indirect
	darvaza.org/slog/handlers/cblog v0.5.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)
