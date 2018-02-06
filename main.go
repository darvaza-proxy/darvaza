package main

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	logger *gnoccoLogger
	//Version contains the git hashtag injected by make
	Version = "N/A"
	//BuildTime contains the build timestamp injected by make
	BuildTime = "N/A"
)

func main() {
	cf := loadConfig()
	initLogger()

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
			logger.fatal("Got SIGTERM, stoping as requested")
		case syscall.SIGINT:
			aserver.shutDown()
			logger.fatal("Got SIGINT, stoping as requested")
		case syscall.SIGUSR2:
			logger.info("Got SIGUSR2, dumping cache")
			aserver.dumpCache()
		default:
			logger.warn("I received %s signal", sign)
		}
	}
}

func initLogger() {
	logger = newLogger()

	if mainconfig.Log.Stdout {
		logger.setLogger("console", nil)
	}

	if mainconfig.Log.File != "" {
		cfg := map[string]interface{}{"file": mainconfig.Log.File}
		logger.setLogger("file", cfg)
		logger.info("Logger started")
	}

}
