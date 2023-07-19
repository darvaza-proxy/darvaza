module darvaza.org/darvaza/server

go 1.19

replace darvaza.org/darvaza/shared => ../shared

require (
	darvaza.org/darvaza/shared v0.5.3
	darvaza.org/slog v0.5.2
	github.com/miekg/dns v1.1.55
	github.com/naoina/toml v0.1.1
)

require (
	darvaza.org/core v0.9.4 // indirect
	darvaza.org/slog/handlers/cblog v0.5.4 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/tools v0.11.0 // indirect
)
