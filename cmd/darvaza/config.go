package main

import (
	"io"
	"log"

	"github.com/creasty/defaults"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Config struct{}

func NewConfig() *Config {
	c := &Config{}

	if err := defaults.Set(c); err != nil {
		log.Fatal(err)
	}

	return c
}

func (c *Config) ReadInFile(filename string) error {
	return hclsimple.DecodeFile(filename, nil, c)
}

func (c *Config) WriteTo(out io.Writer) (int64, error) {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return f.WriteTo(out)
}
