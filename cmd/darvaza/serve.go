package main

import (
	"github.com/spf13/cobra"
)

// Command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts serving a proxy",
	Run: func(cmd *cobra.Command, args []string) {
		cfg.RunProxies()
	},
}

// Flags
func init() {
	rootCmd.AddCommand(serveCmd)
}
