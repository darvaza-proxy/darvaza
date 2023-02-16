module github.com/darvaza-proxy/darvaza/server

go 1.18

replace github.com/darvaza-proxy/darvaza/shared => ../shared

require (
	github.com/darvaza-proxy/darvaza/shared v0.0.0-20230114222335-0836b73ac9de
	github.com/darvaza-proxy/slog v0.2.0
	github.com/miekg/dns v1.1.50
	github.com/naoina/toml v0.1.1
)

require (
	github.com/darvaza-proxy/slog/handlers/cblog v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/tools v0.5.0 // indirect
)
