package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	omnidevx "github.com/plexusone/omnidevx-core"
	report "github.com/plexusone/omnidevx-core/report"
	"github.com/plexusone/omnidevx-core/store"

	"github.com/plexusone/devfolio/output/devxdashboard"
)

var (
	devxDashboardPerson   string
	devxDashboardDays     int
	devxDashboardStoreDir string
	devxDashboardOutput   string
)

var devxDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Export a dashforge dashboard from the local OmniDevX store",
	Long: `Build a DeveloperPeriodReport from locally collected OmniDevX events
and export it as a dashforge Dashboard JSON file: headline metric tiles,
daily activity/cost charts, and a source-coverage table.

The dashboard is a single portable JSON file — open it in dashforge's
static viewer (viewer/index.html?dashboard=<file>) or validate it with
"dashforge validate <file>".

Examples:
  # Last 30 days, written to stdout
  devfolio devx dashboard --person person:jane

  # Last 7 days, written to a file
  devfolio devx dashboard --person person:jane --days 7 -o dashboard.json`,
	RunE: runDevxDashboard,
}

func init() {
	devxDashboardCmd.Flags().StringVar(&devxDashboardPerson, "person", "", "Canonical personId to report on (required)")
	devxDashboardCmd.Flags().IntVar(&devxDashboardDays, "days", 30, "Number of days ending today to report on")
	devxDashboardCmd.Flags().StringVar(&devxDashboardStoreDir, "store-dir", "", "OmniDevX store directory (default: ~/.plexusone/omnidevx/data)")
	devxDashboardCmd.Flags().StringVarP(&devxDashboardOutput, "output", "o", "", "Output file (default: stdout)")
	_ = devxDashboardCmd.MarkFlagRequired("person")
	devxCmd.AddCommand(devxDashboardCmd)
}

func runDevxDashboard(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	if devxDashboardDays <= 0 {
		return fmt.Errorf("--days must be positive")
	}

	s, err := store.Open(store.Options{Dir: devxDashboardStoreDir})
	if err != nil {
		return fmt.Errorf("opening omnidevx store: %w", err)
	}

	end := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.Add(-time.Duration(devxDashboardDays) * 24 * time.Hour)
	period := omnidevx.Period{Start: start, End: end}

	read, err := s.Read(ctx, store.Query{Period: period})
	if err != nil {
		return fmt.Errorf("reading omnidevx store: %w", err)
	}
	for _, d := range read.Diagnostics {
		fmt.Fprintf(os.Stderr, "warning: %s: %s\n", d.Path, d.Message)
	}

	r := report.Build(read.Events, report.Subject{PersonID: devxDashboardPerson}, period)
	daily := buildDailySeries(read.Events, period)

	dash, err := devxdashboard.Export(r, daily)
	if err != nil {
		return fmt.Errorf("exporting dashboard: %w", err)
	}

	output, err := json.MarshalIndent(dash, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling dashboard: %w", err)
	}

	if devxDashboardOutput == "" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(devxDashboardOutput, output, 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote dashboard to %s\n", devxDashboardOutput)
	}

	fmt.Fprintf(os.Stderr, "\nDevX dashboard summary:\n")
	fmt.Fprintf(os.Stderr, "  Person:    %s\n", devxDashboardPerson)
	fmt.Fprintf(os.Stderr, "  Period:    %s .. %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	fmt.Fprintf(os.Stderr, "  Events:    %d\n", len(read.Events))
	fmt.Fprintf(os.Stderr, "  Sources:   %d\n", len(r.Sources))
	fmt.Fprintf(os.Stderr, "  Coverage:  %.0f%%\n", r.Quality.CoverageScore*100)
	if len(r.Quality.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "  Warnings:  %d\n", len(r.Quality.Warnings))
	}

	return nil
}

// buildDailySeries reduces events into one devxdashboard.DailyPoint per
// calendar day in the period, using report.BuildDaily so the chart data
// stays consistent with how the period report itself is aggregated.
func buildDailySeries(events []omnidevx.Event, period omnidevx.Period) []devxdashboard.DailyPoint {
	byDay := map[time.Time][]omnidevx.Event{}
	for _, e := range events {
		if !period.Contains(e.Timestamp) {
			continue
		}
		day := e.Timestamp.UTC().Truncate(24 * time.Hour)
		byDay[day] = append(byDay[day], e)
	}

	days := make([]time.Time, 0, len(byDay))
	for d := range byDay {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	points := make([]devxdashboard.DailyPoint, 0, len(days))
	for _, d := range days {
		daily := report.BuildDaily(byDay[d], d)
		points = append(points, devxdashboard.DailyPoint{
			Date:    d.Format("2006-01-02"),
			Commits: daily.Combined["commits"],
			Prompts: daily.Combined["prompts"],
			CostUSD: daily.Combined["cost_usd"],
		})
	}
	return points
}
