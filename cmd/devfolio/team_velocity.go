package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/grokify/structured-changelog/aggregate"
)

var (
	teamVelocityOutput      string
	teamVelocityGranularity string
	teamVelocitySince       string
	teamVelocityUntil       string
)

var teamVelocityCmd = &cobra.Command{
	Use:   "velocity <portfolio.json>",
	Short: "Generate team velocity dashboard",
	Long: `Generate team velocity metrics and dashboard from aggregated changelog data.

The portfolio file should be generated using 'schangelog portfolio aggregate'.

Metrics include:
  - Total releases and changelog entries
  - Breakdown by category (features, fixes, improvements, etc.)
  - Time series data for velocity trends
  - Activity heatmap data (GitHub-style)
  - Per-project contribution breakdown

Examples:
  # Generate velocity dashboard
  devfolio team velocity portfolio.json -o velocity.json

  # Filter by date range
  devfolio team velocity portfolio.json --since 2024-01-01 --until 2024-06-30`,
	Args: cobra.ExactArgs(1),
	RunE: runTeamVelocity,
}

func init() {
	teamVelocityCmd.Flags().StringVarP(&teamVelocityOutput, "output", "o", "", "Output file (default: stdout)")
	teamVelocityCmd.Flags().StringVar(&teamVelocityGranularity, "granularity", "week", "Time granularity: day, week, month")
	teamVelocityCmd.Flags().StringVar(&teamVelocitySince, "since", "", "Start date (YYYY-MM-DD)")
	teamVelocityCmd.Flags().StringVar(&teamVelocityUntil, "until", "", "End date (YYYY-MM-DD)")
	teamCmd.AddCommand(teamVelocityCmd)
}

func runTeamVelocity(cmd *cobra.Command, args []string) error {
	portfolioPath := args[0]

	// Load portfolio using structured-changelog's aggregate package
	portfolio, err := aggregate.LoadPortfolioFile(portfolioPath)
	if err != nil {
		return fmt.Errorf("loading portfolio: %w", err)
	}

	// Build metrics options
	opts := aggregate.MetricsOptions{
		Granularity:    teamVelocityGranularity,
		IncludeRollups: true,
	}

	if teamVelocitySince != "" {
		t, err := time.Parse("2006-01-02", teamVelocitySince)
		if err != nil {
			return fmt.Errorf("invalid --since date: %w", err)
		}
		opts.Since = t
	}

	if teamVelocityUntil != "" {
		t, err := time.Parse("2006-01-02", teamVelocityUntil)
		if err != nil {
			return fmt.Errorf("invalid --until date: %w", err)
		}
		opts.Until = t
	}

	// Calculate metrics
	metrics, err := aggregate.CalculateMetrics(portfolio, opts)
	if err != nil {
		return fmt.Errorf("calculating metrics: %w", err)
	}

	// Export dashboard data
	export, err := aggregate.ExportDashboard(metrics)
	if err != nil {
		return fmt.Errorf("exporting dashboard: %w", err)
	}

	// Generate dashboard JSON
	dashOpts := aggregate.DefaultDashboardOptions()
	dashOpts.Title = portfolio.Name + " - Team Velocity"
	dashOpts.Template = "velocity"

	output, err := aggregate.GenerateDashboardJSON(export, dashOpts)
	if err != nil {
		return fmt.Errorf("generating dashboard: %w", err)
	}

	// Write output
	if teamVelocityOutput == "" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(teamVelocityOutput, output, 0600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote team velocity dashboard to %s\n", teamVelocityOutput)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\nTeam velocity summary:\n")
	fmt.Fprintf(os.Stderr, "  Projects:  %d\n", len(metrics.ByProject))
	fmt.Fprintf(os.Stderr, "  Releases:  %d\n", metrics.TotalReleases)
	fmt.Fprintf(os.Stderr, "  Changes:   %d\n", metrics.TotalEntries)

	if len(metrics.ByRollup) > 0 {
		fmt.Fprintf(os.Stderr, "  By type:\n")
		for name, count := range metrics.ByRollup {
			fmt.Fprintf(os.Stderr, "    %s: %d\n", name, count)
		}
	}

	return nil
}

// Blank identifier to ensure json import is used (for future expansion)
var _ = json.Marshal
