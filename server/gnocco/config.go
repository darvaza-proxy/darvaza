package gnocco

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/naoina/toml"

	"github.com/darvaza-proxy/slog"
)

// Gnocco is the configuration representing the dns-resolver
type Gnocco struct {
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

	hosts  hostsCfg
	logger slog.Logger
}

type hostsCfg struct {
	Enable          bool
	HostsFile       string
	RefreshInterval uint
}

func NewFromFilename(f string) (*Gnocco, error) {
	var cf Gnocco

	if f == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("error %s occurred", err)
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
		return nil, fmt.Errorf("error %s occurred", err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error %s occurred", err)
	}

	if err := toml.Unmarshal(buf, &cf); err != nil {
		return nil, fmt.Errorf("error %s occurred", err)
	}

	cf.logger = newLogger(&cf)
	return &cf, nil
}
