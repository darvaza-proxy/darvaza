// Gnocco is a little cache of goodness
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

var mainconfig cfg

func loadConfig(f string) (cfg, error) {

	if f == "" {
		ex, err := os.Executable()
		if err != nil {
			return mainconfig, fmt.Errorf("error %s occurred", err)
		}
		confPath := filepath.Dir(ex) + "/gnocco.conf"
		if _, err := os.Stat("/etc/gnocco/gnocco.conf"); err == nil {
			f = "/etc/gnocco/gnocco.conf"
		}
		if _, err := os.Stat(confPath); err == nil {
			f = confPath
		}
	}
	file, err := os.Open(f)
	if err != nil {
		return mainconfig, fmt.Errorf("error %s occurred", err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return mainconfig, fmt.Errorf("error %s occurred", err)
	}

	if err := toml.Unmarshal(buf, &mainconfig); err != nil {
		return mainconfig, fmt.Errorf("error %s occurred", err)
	}
	return mainconfig, nil
}
