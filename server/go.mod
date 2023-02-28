module github.com/darvaza-proxy/darvaza/server

go 1.18

replace github.com/darvaza-proxy/darvaza/shared => ../shared

require (
	github.com/darvaza-proxy/darvaza/shared v0.1.3
	github.com/darvaza-proxy/slog v0.4.4
	github.com/miekg/dns v1.1.51
	github.com/naoina/toml v0.1.1
)

require (
	github.com/darvaza-proxy/core v0.1.3 // indirect
	github.com/darvaza-proxy/slog/handlers/cblog v0.2.2 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
)
