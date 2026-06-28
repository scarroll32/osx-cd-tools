package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cd-tools",
	Short: "macOS CD utilities",
	Long:  "cd-tools provides utilities for working with audio CDs on macOS.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
