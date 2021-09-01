package main

import (
	"os"

	"github.com/spf13/cobra"
)

// Command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "prints the config",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := cfg.WriteTo(os.Stdout)
		return err
	},
}

// Flags
func init() {
	rootCmd.AddCommand(dumpCmd)
}
