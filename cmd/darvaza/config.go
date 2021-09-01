package main

import (
	"log"

	"github.com/creasty/defaults"
	"github.com/hashicorp/hcl/v2/hclsimple"
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
