module github.com/darvaza-proxy/darvaza

go 1.19

require (
	github.com/creasty/defaults v1.6.0
	github.com/darvaza-proxy/darvaza/server v0.0.1
	github.com/darvaza-proxy/darvaza/shared v0.2.1
	github.com/hashicorp/hcl/v2 v2.16.1
	github.com/mgechev/revive v1.2.5
	github.com/miekg/dns v1.1.51
	github.com/spf13/cobra v1.6.1
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/chavacava/garif v0.0.0-20230227094218-b8c73b2037b8 // indirect
	github.com/darvaza-proxy/core v0.5.2 // indirect
	github.com/darvaza-proxy/slog v0.4.5 // indirect
	github.com/darvaza-proxy/slog/handlers/cblog v0.4.0 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mgechev/dots v0.0.0-20210922191527-e955255bf517 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zclconf/go-cty v1.13.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
)

replace (
	github.com/darvaza-proxy/darvaza/acme => ./acme
	github.com/darvaza-proxy/darvaza/agent => ./agent
	github.com/darvaza-proxy/darvaza/server => ./server
	github.com/darvaza-proxy/darvaza/shared => ./shared
)
