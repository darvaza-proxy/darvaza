package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/karasz/gnocco/cblog"
)

//go:generate go run genroot.go

var (
	logger *log.Logger
	//Version contains the git hashtag injected by make
	Version = "N/A"
	//BuildTime contains the build timestamp injected by make
	BuildTime = "N/A"
)

func main() {
	cf, err := loadConfig()
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
		default:
			logger.Warn("I received %s signal", sign)
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
