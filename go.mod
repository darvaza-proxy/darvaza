module darvaza.org/darvaza

go 1.21

require (
	darvaza.org/core v0.14.8
	darvaza.org/darvaza/server v0.1.3
	darvaza.org/darvaza/shared v0.5.12
	darvaza.org/darvaza/shared/config v0.2.6
	darvaza.org/slog v0.5.10 // indirect
	darvaza.org/slog/handlers/cblog v0.5.11 // indirect
	darvaza.org/x/tls v0.2.0 // indirect
)

require (
	github.com/hashicorp/hcl/v2 v2.19.1
	github.com/mgechev/revive v1.3.4
	github.com/miekg/dns v1.1.57
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/amery/defaults v0.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/chavacava/garif v0.1.0 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mgechev/dots v0.0.0-20210922191527-e955255bf517 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zclconf/go-cty v1.14.1 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)

replace (
	darvaza.org/darvaza/acme => ./acme
	darvaza.org/darvaza/agent => ./agent
	darvaza.org/darvaza/server => ./server
	darvaza.org/darvaza/shared => ./shared
)
