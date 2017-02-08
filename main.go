package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		logger.Info("signal %s received", sig)
		done <- true
	}()
	<-done
	logger.Info("stopping")
}

func initLogger() {
	logger = NewLogger()

	if Config.Log.Stdout {
		logger.SetLogger("console", nil)
	}

	if Config.Log.File != "" {
		cfg := map[string]interface{}{"file": Config.Log.File}
		logger.SetLogger("file", cfg)
		logger.Info("Logger started")
	}

}
