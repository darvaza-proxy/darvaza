module github.com/darvaza-proxy/darvaza/server

go 1.19

replace github.com/darvaza-proxy/darvaza/shared => ../shared

require (
	github.com/darvaza-proxy/darvaza/shared v0.2.1
	github.com/darvaza-proxy/slog v0.4.5
	github.com/miekg/dns v1.1.51
	github.com/naoina/toml v0.1.1
)

require (
	github.com/darvaza-proxy/core v0.5.2 // indirect
	github.com/darvaza-proxy/slog/handlers/cblog v0.4.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
)
