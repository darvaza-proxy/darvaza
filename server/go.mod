module github.com/darvaza-proxy/darvaza/server

go 1.19

replace github.com/darvaza-proxy/darvaza/shared => ../shared

require (
	darvaza.org/slog v0.5.0
	github.com/darvaza-proxy/darvaza/shared v0.4.3
	github.com/miekg/dns v1.1.52
	github.com/naoina/toml v0.1.1
)

require (
	darvaza.org/core v0.9.0 // indirect
	darvaza.org/slog/handlers/cblog v0.5.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
)
