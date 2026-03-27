module darvaza.org/darvaza/agent

go 1.24

require (
	darvaza.org/core v0.16.1
	darvaza.org/darvaza/acme v0.3.0
	darvaza.org/middleware v0.3.1
	darvaza.org/slog v0.6.1
	darvaza.org/slog/handlers/discard v0.5.1
	darvaza.org/x/fs v0.4.1 // indirect
	darvaza.org/x/net v0.5.1
	darvaza.org/x/tls v0.5.1
	darvaza.org/x/web v0.10.0 // indirect
)

require (
	github.com/quic-go/quic-go v0.59.0
	golang.org/x/net v0.43.0
)

require (
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)
