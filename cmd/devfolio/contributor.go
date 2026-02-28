package main

import (
	"github.com/spf13/cobra"
)

var contributorCmd = &cobra.Command{
	Use:     "contributor",
	Aliases: []string{"contrib"},
	Short:   "Individual contributor portfolio and metrics",
	Long: `Generate individual contributor portfolios from changelog data and GitHub activity.

Subcommands:
  profile    Generate contributor profile/portfolio
  history    View contribution history over time
  export     Export portfolio data

Use cases:
  - Self-reflection: Track your own contributions over time
  - Hiring managers: Evaluate candidate contribution patterns
  - Job seekers: Generate portfolio showcasing your work

Examples:
  # Generate your contributor profile
  devfolio contributor profile --user grokify -o my-portfolio.json

  # View contribution history for a user
  devfolio contributor history --user grokify --since 2024-01-01`,
}

func init() {
	rootCmd.AddCommand(contributorCmd)
}
