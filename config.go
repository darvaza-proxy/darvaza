package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

type cfg struct {
	User           string
	Group          string
	RootsFile      string
	PermissionsDir string
	Daemon         bool
	DoTCP          bool

	Listen struct {
		Host string
		Port int
	}

	Log struct {
		Stdout bool
		File   string
	}

	Hosts HostsCfg
}

type HostsCfg struct {
	Enable           bool
	Hosts_File       string
	Refresh_Interval uint
}

var (
	Config   cfg
	confFile string
)

func loadConfig() cfg {
	flag.StringVar(&confFile, "f", "/etc/gnocco/gnocco.conf", "specify the config file, defaults to /etc/gnocco/gnocco.conf.")

	flag.Parse()

	file, err := os.Open(confFile)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if err := toml.Unmarshal(buf, &Config); err != nil {
		panic(err)
	}
	return Config
}
