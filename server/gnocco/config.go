package gnocco

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/naoina/toml"

	"github.com/darvaza-proxy/slog"
)

// ListenCfg is a host:port combination
type ListenCfg struct {
	Host string
	Port int
}

// LogCfg is the configuration for logging
type LogCfg struct {
	Stdout bool
	File   string
}

// CacheCfg is the configuration for the cache
type CacheCfg struct {
	DumpInterval int
	Expire       int
	MaxCount     int
	CachePath    string
}

// Gnocco is the configuration representing the dns-resolver
type Gnocco struct {
	RootsFile      string
	PermissionsDir string
	Daemon         bool
	IterateResolv  bool
	Listen         ListenCfg
	MaxJobs        int
	MaxQueries     int
	Log            LogCfg
	Cache          CacheCfg
	hosts          HostsCfg
	logger         slog.Logger
}

// HostsCfg is the configuration for using hosts file
type HostsCfg struct {
	Enable          bool
	HostsFile       string
	RefreshInterval uint
}

// NewFromTOML creates a new Gnocco configuration from a TOML file
func NewFromTOML(f string) (*Gnocco, error) {
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
