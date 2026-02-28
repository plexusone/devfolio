package main

import (
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Team velocity metrics and dashboards",
	Long: `Generate team velocity metrics from aggregated changelog data.

Subcommands:
  velocity   Generate team velocity dashboard
  compare    Compare velocity across time periods
  export     Export metrics data

Examples:
  # Generate team velocity from a portfolio
  devfolio team velocity portfolio.json -o team-dashboard.json

  # Compare Q1 vs Q2 velocity
  devfolio team compare portfolio.json --period1 2024-Q1 --period2 2024-Q2`,
}

func init() {
	rootCmd.AddCommand(teamCmd)
}
