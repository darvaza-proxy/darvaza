package main

import (
	"github.com/spf13/cobra"
)

// Command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts serving a proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := newServer()
		return srv.Run()
	},
}

// Flags
func init() {
	rootCmd.AddCommand(serveCmd)
}
