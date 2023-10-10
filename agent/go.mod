module darvaza.org/darvaza/agent

go 1.19

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.10.0
	darvaza.org/darvaza/acme v0.1.7
	darvaza.org/darvaza/shared v0.5.11
	darvaza.org/middleware v0.2.6
	darvaza.org/slog v0.5.4
	darvaza.org/slog/handlers/discard v0.4.6
	github.com/quic-go/quic-go v0.39.0
	golang.org/x/net v0.17.0
)

require (
	darvaza.org/darvaza/shared/web v0.3.11 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20230926050212-f7f687d19a98 // indirect
	github.com/onsi/ginkgo/v2 v2.13.0 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.3.4 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/mock v0.3.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
