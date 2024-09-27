package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	darvaza "darvaza.org/darvaza/server"
)

// Command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts serving a proxy",
	RunE: func(_ *cobra.Command, _ []string) error {
		server := darvaza.NewServer()
		for i := range cfg.Proxies {
			if z := cfg.Proxies[i].New(); z != nil {
				server.Append(z)
			}
		}

		go func() {
			_ = server.Run()
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		defer close(sig)

		for {
			select {
			case signum := <-sig:
				switch signum {
				case syscall.SIGHUP:
					log.Println("Reloading")
					err := server.Reload()
					if err != nil {
						log.Println(err)
						return err
					}
				case syscall.SIGINT, syscall.SIGTERM:
					log.Println("Terminating")
					err := server.Cancel()
					if err != nil {
						log.Println(err)
						return err
					}
				}
			case err := <-server.Done:
				return err
			}
		}
	},
}

// Flags
func init() {
	rootCmd.AddCommand(serveCmd)
}
