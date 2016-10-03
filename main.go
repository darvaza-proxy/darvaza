package main

import (
	"os"
	"os/signal"
)

var (
	logger *GnoccoLogger
)

func main() {

	initLogger()

	server := &Server{
		thost: Config.TCPServer.Host,
		tport: Config.TCPServer.Port,
		uhost: Config.UDPServer.Host,
		uport: Config.UDPServer.Port,
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
