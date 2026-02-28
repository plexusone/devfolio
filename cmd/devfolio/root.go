package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devfolio",
	Short: "Developer portfolio and team velocity metrics",
	Long: `devfolio generates developer portfolios and team velocity dashboards
from changelog data, git history, and GitHub activity.

Use cases:
  - Team velocity dashboards for engineering managers
  - Individual contributor self-reflection and growth tracking
  - Recruiting: evaluate candidates' contribution history
  - Job seeking: showcase your development portfolio

Workflow:
  1. Configure sources:  devfolio init -o devfolio.json
  2. Generate portfolio: devfolio generate devfolio.json -o portfolio/
  3. View dashboard:     devfolio serve portfolio/`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
