module darvaza.org/darvaza/server

go 1.19

replace darvaza.org/darvaza/shared => ../shared

require (
	darvaza.org/darvaza/shared v0.5.10
	darvaza.org/slog v0.5.4
	github.com/miekg/dns v1.1.56
	github.com/naoina/toml v0.1.1
)

require (
	darvaza.org/core v0.11.0 // indirect
	darvaza.org/slog/handlers/cblog v0.5.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.0 // indirect
)
