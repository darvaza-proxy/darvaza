// Package config contains funtions to deal with TLSproxy configs
package main

import (
	"io"
	"log"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"darvaza.org/darvaza/shared/config"
	"darvaza.org/darvaza/shared/tls/server"
)

// Config is the main configuration item containig all
// the ProxyConfigs.
type Config struct {
	Proxies []server.ProxyConfig `hcl:"proxy,block"`
}

// SetDefaults is calling Set to set the default values
// in a Config.
func (c *Config) SetDefaults() {
	defaultProxy := &server.ProxyConfig{}
	if err := config.SetDefaults(defaultProxy); err != nil {
		log.Println(err)
	}
	if config.CanUpdate(c.Proxies) {
		c.Proxies = append(c.Proxies, *defaultProxy)
	}
}

// NewConfig returns a pointer to a Config with
// defaults set
func NewConfig() *Config {
	c := &Config{}

	if err := config.SetDefaults(c); err != nil {
		// revive:disable:deep-exit
		log.Fatal(err)
		// revive:enable:deep-exit
	}

	return c
}

// ReadInFile fills a Config with values from a hcl file.
func (c *Config) ReadInFile(filename string) error {
	return hclsimple.DecodeFile(filename, nil, c)
}

// WriteTo writes a Config to a file in hcl format
func (c *Config) WriteTo(out io.Writer) (int64, error) {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return f.WriteTo(out)
}
