package main

import (
	"sync"

	"github.com/spf13/cobra"
)

// Command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts serving a proxy",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		for i := range cfg.Proxies {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cfg.Proxies[i].Run()
			}(i)
		}
		wg.Wait()
	},
}

// Flags
func init() {
	rootCmd.AddCommand(serveCmd)
}
