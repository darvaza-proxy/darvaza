module darvaza.org/darvaza

go 1.22.0

toolchain go1.22.10

require (
	darvaza.org/core v0.16.1
	darvaza.org/darvaza/server v0.2.0
	darvaza.org/darvaza/shared v0.7.0
	darvaza.org/slog v0.6.0 // indirect
	darvaza.org/slog/handlers/cblog v0.6.0 // indirect
	darvaza.org/x/config v0.4.2
	darvaza.org/x/tls v0.5.0 // indirect
)

require (
	github.com/hashicorp/hcl/v2 v2.23.0
	github.com/miekg/dns v1.1.63
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/amery/defaults v0.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.25.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zclconf/go-cty v1.16.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ./acme
	darvaza.org/darvaza/agent => ./agent
	darvaza.org/darvaza/server => ./server
	darvaza.org/darvaza/shared => ./shared
)
