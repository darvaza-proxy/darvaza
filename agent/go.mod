module github.com/darvaza-proxy/darvaza/agent

go 1.19

replace (
	github.com/darvaza-proxy/darvaza/acme => ../acme
	github.com/darvaza-proxy/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.9.0
	darvaza.org/slog v0.5.0
	darvaza.org/slog/handlers/discard v0.4.0
	github.com/darvaza-proxy/darvaza/acme v0.0.4
	github.com/darvaza-proxy/darvaza/shared v0.4.6
	github.com/darvaza-proxy/middleware v0.0.5
	github.com/quic-go/quic-go v0.33.0
	golang.org/x/net v0.8.0
)

require (
	github.com/darvaza-proxy/core v0.6.5 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20230309165930-d61513b1440d // indirect
	github.com/onsi/ginkgo/v2 v2.9.1 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-19 v0.3.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.2.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
)
