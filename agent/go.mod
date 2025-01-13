module darvaza.org/darvaza/agent

go 1.22

require (
	darvaza.org/core v0.16.0
	darvaza.org/darvaza/acme v0.2.0
	darvaza.org/middleware v0.3.1
	darvaza.org/slog v0.6.0
	darvaza.org/slog/handlers/discard v0.5.0
	darvaza.org/x/fs v0.3.8 // indirect
	darvaza.org/x/net v0.4.4
	darvaza.org/x/tls v0.4.5
	darvaza.org/x/web v0.9.6 // indirect
)

require (
	github.com/quic-go/quic-go v0.48.2
	golang.org/x/net v0.34.0
)

require (
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20231229205709-960ae82b1e42 // indirect
	github.com/onsi/ginkgo/v2 v2.15.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)
