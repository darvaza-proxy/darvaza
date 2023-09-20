module darvaza.org/darvaza/agent

go 1.19

replace (
	darvaza.org/darvaza/acme => ../acme
	darvaza.org/darvaza/shared => ../shared
)

require (
	darvaza.org/core v0.9.9
	darvaza.org/darvaza/acme v0.1.4
	darvaza.org/darvaza/shared v0.5.9
	darvaza.org/middleware v0.2.5
	darvaza.org/slog v0.5.4
	darvaza.org/slog/handlers/discard v0.4.6
	github.com/quic-go/quic-go v0.38.1
	golang.org/x/net v0.15.0
)

require (
	darvaza.org/darvaza/shared/web v0.3.10 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20230912144702-c363fe2c2ed8 // indirect
	github.com/onsi/ginkgo/v2 v2.12.1 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.3.4 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
