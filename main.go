// Gnocco is a little cache of goodness
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/darvaza-proxy/gnocco/cblog"
)

//go:generate go run genroot.go

var (
	logger *log.Logger
	//Version contains the git hashtag injected by make
	Version = "N/A"
	//BuildDate contains the build timestamp injected by make
	BuildDate = "N/A"
)

func main() {
	var confFile string
	var vrs bool
	flag.StringVar(&confFile, "f", "", "specify the config file, if empty will try gnocco.conf and /etc/gnocco/gnocco.conf.")
	flag.BoolVar(&vrs, "v", false, "program version")
	flag.Parse()

	if vrs {
		fmt.Fprintf(os.Stdout, "Gnocco version %s, build date %s\n", Version, BuildDate)
		os.Exit(0)
	}

	cf, err := loadConfig(confFile)
	if err != nil {
		panic(err)
	}
	logger = initLogger()

	aserver := &server{
		host:       cf.Listen.Host,
		port:       cf.Listen.Port,
		maxjobs:    cf.MaxJobs,
		maxqueries: cf.MaxQueries,
	}

	aserver.run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)

	for {
		sign := <-signalChan
		switch sign {
		case syscall.SIGTERM:
			aserver.shutDown()
			logger.Fatal("Got SIGTERM, stoping as requested")
		case syscall.SIGINT:
			aserver.shutDown()
			logger.Fatal("Got SIGINT, stoping as requested")
		case syscall.SIGUSR2:
			logger.Info("Got SIGUSR2, dumping cache")
			aserver.dumpCache()
		case syscall.SIGURG:
		default:
			logger.Warn("I received %v signal", sign)
		}
	}
}

func initLogger() *log.Logger {
	logger = log.New()

	if mainconfig.Log.Stdout {
		logger.SetLogger("console", nil)
	}

	if mainconfig.Log.File != "" {
		cfg := map[string]interface{}{"file": mainconfig.Log.File}
		logger.SetLogger("file", cfg)
		logger.Info("Logger started")
	}
	return logger
}
