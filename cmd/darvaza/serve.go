package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

// Command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts serving a proxy",
	Run: func(cmd *cobra.Command, args []string) {
		startProxies()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		defer close(sig)
	out:
		for {
			select {
			case signum := <-sig:
				switch signum {
				case syscall.SIGHUP:
					log.Println("Reloading")
					reloadProxies()
				case syscall.SIGINT, syscall.SIGTERM:
					log.Println("Terminating")
					stopProxies()
					// TODO: replace the break out with a return while returning
					// this to a RunE
					break out
				}
			}
		}
	},
}

func reloadProxies() {
	for i := range cfg.Proxies {
		err := cfg.Proxies[i].Reload()
		if err != nil {
			log.Println(err)
		}
	}
}

func startProxies() {
	var wg sync.WaitGroup
	for i := range cfg.Proxies {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cfg.Proxies[i].Run()
		}(i)
	}
	wg.Wait()
}

func stopProxies() {
	for i := range cfg.Proxies {
		err := cfg.Proxies[i].Cancel()
		if err != nil {
			log.Println(err)
		}
	}
}

// Flags
func init() {
	rootCmd.AddCommand(serveCmd)
}
