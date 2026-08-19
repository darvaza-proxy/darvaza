module darvaza.org/darvaza/server

go 1.25.0

require (
	darvaza.org/core v0.16.1
	darvaza.org/darvaza/shared v0.7.0
	darvaza.org/slog v0.6.1
	darvaza.org/slog/handlers/cblog v0.6.1 // indirect
)

require (
	github.com/miekg/dns v1.1.73
	github.com/naoina/toml v0.1.1
)

require (
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace darvaza.org/darvaza/shared => ../shared
