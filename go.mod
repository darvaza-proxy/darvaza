module darvaza.org/darvaza

go 1.21

require (
	darvaza.org/core v0.15.3
	darvaza.org/darvaza/server v0.1.5
	darvaza.org/darvaza/shared v0.6.2
	darvaza.org/slog v0.5.14 // indirect
	darvaza.org/slog/handlers/cblog v0.5.13 // indirect
	darvaza.org/x/config v0.3.8
	darvaza.org/x/tls v0.3.0 // indirect
)

require (
	github.com/hashicorp/hcl/v2 v2.19.1
	github.com/miekg/dns v1.1.62
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/amery/defaults v0.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/zclconf/go-cty v1.14.1 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ./acme
	darvaza.org/darvaza/agent => ./agent
	darvaza.org/darvaza/server => ./server
	darvaza.org/darvaza/shared => ./shared
)
