package main

import (
	"os"
	"os/signal"
)

var (
	logger *GnoccoLogger
)

func main() {
	cf := loadConfig()
	initLogger()
	initResolver()

	server := &Server{
		thost: cf.TCPServer.Host,
		tport: cf.TCPServer.Port,
		uhost: cf.UDPServer.Host,
		uport: cf.UDPServer.Port,
		user:  cf.User,
		group: cf.Group,
	}

	server.Run()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			logger.Info("signal received, stopping")
			break forever
		}
	}

}

func initLogger() {
	logger = NewLogger()

	if Config.Log.Stdout {
		logger.SetLogger("console", nil)
	}

	if Config.Log.File != "" {
		cfg := map[string]interface{}{"file": Config.Log.File}
		logger.SetLogger("file", cfg)
	}

}
