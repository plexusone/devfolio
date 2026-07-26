package main

import (
	"github.com/spf13/cobra"
)

var quarterlyCmd = &cobra.Command{
	Use:   "quarterly",
	Short: "Quarterly developer reports",
	Long:  `Generate quarterly developer reports from GitHub, git, changelog, and token spend data.`,
}

func init() {
	rootCmd.AddCommand(quarterlyCmd)
}
