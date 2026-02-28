package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set by goreleaser at build time.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devfolio %s\n", Version)
	},
}
