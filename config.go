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
	SafeResolver   bool

	Listen struct {
		Host string
		Port int
	}

	Log struct {
		Stdout bool
		File   string
	}

	Cache struct {
		Expire   int
		MaxCount int
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
		logger.Fatal(err)
	}

	if err := toml.Unmarshal(buf, &Config); err != nil {
		logger.Fatal(err)
	}
	return Config
}
