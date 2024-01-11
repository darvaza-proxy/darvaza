module darvaza.org/darvaza/agent

go 1.19

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.11.3
	darvaza.org/darvaza/acme v0.1.8
	darvaza.org/darvaza/shared v0.5.12
	darvaza.org/middleware v0.2.7
	darvaza.org/slog v0.5.6
	darvaza.org/slog/handlers/discard v0.4.9
	github.com/quic-go/quic-go v0.40.1
	golang.org/x/net v0.20.0
)

require (
	darvaza.org/darvaza/shared/web v0.3.12 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20231229205709-960ae82b1e42 // indirect
	github.com/onsi/ginkgo/v2 v2.13.2 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/exp v0.0.0-20240110193028-0dcbfd608b1e // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
