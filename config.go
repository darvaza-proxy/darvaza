package main

import (
	"flag"
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

	TCPServer struct {
		Host string
		Port int
	}

	UDPServer struct {
		Host string
		Port int
	}

	Log struct {
		Stdout bool
		File   string
	}

	HostsCfg HostsConfig
}

type HostsConfig struct {
	Enable           bool
	Hosts_File       string
	Refresh_Interval uint
}

var Config cfg

func init() {
	var confFile string
	flag.StringVar(&confFile, "f", "/etc/gnocco/gnocco.toml", "specify the config file, defaults to /etc/gnocco/gnocco.toml.")

	flag.Parse()

	file, err := os.Open(confFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if err := toml.Unmarshal(buf, &Config); err != nil {
	}
}
