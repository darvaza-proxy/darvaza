module darvaza.org/darvaza/agent

go 1.19

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.9.7
	darvaza.org/darvaza/acme v0.1.2
	darvaza.org/darvaza/shared v0.5.3
	darvaza.org/middleware v0.2.3
	darvaza.org/slog v0.5.3
	darvaza.org/slog/handlers/discard v0.4.5
	github.com/quic-go/quic-go v0.36.2
	golang.org/x/net v0.14.0
)

require (
	darvaza.org/darvaza/shared/web v0.3.7 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20230705174524-200ffdc848b8 // indirect
	github.com/onsi/ginkgo/v2 v2.11.0 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-19 v0.3.2 // indirect
	github.com/quic-go/qtls-go1-20 v0.2.2 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.org/x/tools v0.12.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
