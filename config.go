package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

type cfg struct {
	RootsFile      string
	PermissionsDir string
	Daemon         bool
	IterateResolv  bool

	Listen struct {
		Host string
		Port int
	}

	MaxJobs    int
	MaxQueries int

	Log struct {
		Stdout bool
		File   string
	}

	Cache struct {
		DumpInterval int
		Expire       int
		MaxCount     int
		CachePath    string
	}

	hosts hostsCfg
}

type hostsCfg struct {
	Enable          bool
	HostsFile       string
	RefreshInterval uint
}

var (
	mainconfig cfg
	confFile   string
)

func loadConfig() (cfg, error) {
	flag.StringVar(&confFile, "f", "/etc/gnocco/gnocco.conf", "specify the config file, defaults to /etc/gnocco/gnocco.conf.")

	flag.Parse()

	file, err := os.Open(confFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return mainconfig, fmt.Errorf("Error %s occurred.", err)
	}

	if err := toml.Unmarshal(buf, &mainconfig); err != nil {
		return mainconfig, fmt.Errorf("Error %s occurred.", err)
	}
	return mainconfig, nil
}
