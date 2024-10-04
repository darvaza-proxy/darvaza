module darvaza.org/darvaza/agent

go 1.21

require (
	darvaza.org/core v0.15.0
	darvaza.org/darvaza/acme v0.2.0
	darvaza.org/middleware v0.2.9
	darvaza.org/slog v0.5.11
	darvaza.org/slog/handlers/discard v0.4.14
	darvaza.org/x/fs v0.3.3 // indirect
	darvaza.org/x/net v0.3.4
	darvaza.org/x/tls v0.2.1
	darvaza.org/x/web v0.9.0 // indirect
)

require (
	github.com/quic-go/quic-go v0.40.1
	golang.org/x/net v0.30.0
)

require (
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/pprof v0.0.0-20231229205709-960ae82b1e42 // indirect
	github.com/onsi/ginkgo/v2 v2.13.2 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)
