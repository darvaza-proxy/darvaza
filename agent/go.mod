module darvaza.org/darvaza/agent

go 1.25.0

require (
	darvaza.org/core v0.21.2
	darvaza.org/darvaza/acme v0.3.0
	darvaza.org/middleware v0.3.2
	darvaza.org/slog v0.10.0
	darvaza.org/slog/handlers/discard v0.8.0
	darvaza.org/x/fs v0.7.0 // indirect
	darvaza.org/x/net v0.8.2
	darvaza.org/x/tls v0.9.0
	darvaza.org/x/web v0.10.1 // indirect
)

require (
	github.com/quic-go/quic-go v0.49.0
	golang.org/x/net v0.57.0
)

require (
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	github.com/onsi/ginkgo/v2 v2.22.2 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	go.uber.org/mock v0.5.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)
