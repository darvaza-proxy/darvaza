module darvaza.org/darvaza/agent

go 1.19

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.9.1
	darvaza.org/darvaza/acme v0.1.0
	darvaza.org/darvaza/shared v0.5.0
	darvaza.org/middleware v0.2.1
	darvaza.org/slog v0.5.0
	darvaza.org/slog/handlers/discard v0.4.0
	github.com/quic-go/quic-go v0.33.0
	golang.org/x/net v0.8.0
)

require (
	darvaza.org/darvaza/shared/web v0.3.6 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20230323073829-e72429f035bd // indirect
	github.com/onsi/ginkgo/v2 v2.9.2 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-19 v0.3.2 // indirect
	github.com/quic-go/qtls-go1-20 v0.2.2 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
)
