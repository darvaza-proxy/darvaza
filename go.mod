module github.com/darvaza-proxy/darvaza

go 1.18

require (
	github.com/creasty/defaults v1.6.0
	github.com/darvaza-proxy/darvaza/server v0.0.0-20220815113152-dfebb9141d43
	github.com/darvaza-proxy/darvaza/shared v0.0.0-20220815153810-886eb2ae5f7a
	github.com/hashicorp/hcl/v2 v2.13.0
	github.com/spf13/cobra v1.5.0
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/zclconf/go-cty v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace (
	github.com/darvaza-proxy/darvaza/acme => ./acme
	github.com/darvaza-proxy/darvaza/agent => ./agent
	github.com/darvaza-proxy/darvaza/server => ./server
	github.com/darvaza-proxy/darvaza/shared => ./shared
)
