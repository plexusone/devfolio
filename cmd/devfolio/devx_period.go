package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	omnidevx "github.com/plexusone/omnidevx-core"
	report "github.com/plexusone/omnidevx-core/report"
	"github.com/plexusone/omnidevx-core/store"

	"github.com/plexusone/devfolio/output/devxdashboard"
)

// runDevxPeriodDashboard handles `devfolio devx dashboard --period ...`:
// it resolves the calendar period containing the anchor date, builds a
// devxdashboard.PeriodReport (with weekly/monthly model breakdowns for
// monthly/quarterly reports), and writes the resulting Dashboard IR to the
// standard reports path so the visionstudio daemon's
// /api/devx/periods and /api/devx/reports/{periodType}/{label} endpoints
// can find it.
func runDevxPeriodDashboard() error {
	ctx := context.Background()

	periodType := devxdashboard.PeriodType(devxDashboardPeriod)

	anchor := time.Now()
	if devxDashboardFor != "" {
		parsed, err := time.Parse("2006-01-02", devxDashboardFor)
		if err != nil {
			return fmt.Errorf("parsing --for: %w", err)
		}
		anchor = parsed
	}

	start, end, label, err := periodBoundaries(periodType, anchor)
	if err != nil {
		return err
	}
	period := omnidevx.Period{Start: start, End: end}

	s, err := store.Open(store.Options{Dir: devxDashboardStoreDir})
	if err != nil {
		return fmt.Errorf("opening omnidevx store: %w", err)
	}

	read, err := s.Read(ctx, store.Query{Period: period})
	if err != nil {
		return fmt.Errorf("reading omnidevx store: %w", err)
	}
	for _, d := range read.Diagnostics {
		fmt.Fprintf(os.Stderr, "warning: %s: %s\n", d.Path, d.Message)
	}

	subject := report.Subject{PersonID: devxDashboardPerson}
	r := report.Build(read.Events, subject, period)
	daily := buildDailySeries(read.Events, period)

	pr := &devxdashboard.PeriodReport{
		Type:   periodType,
		Label:  label,
		Report: r,
		Daily:  daily,
	}
	if periodType == devxdashboard.PeriodMonthly || periodType == devxdashboard.PeriodQuarterly {
		// Anchor to the Monday on/before the period start so each bucket is
		// a real calendar week (Mon-Sun), not an arbitrary 7-day chunk from
		// whatever weekday the month/quarter happens to start on — WeekLabel
		// formats the bucket's start date, so misaligned buckets would show
		// as e.g. "Wed Jul 1, 2026" instead of a real week.
		pr.WeeklyByModel = modelPointsForSubPeriods(read.Events, subject, isoWeekStart(start), end,
			func(t time.Time) time.Time { return t.AddDate(0, 0, 7) }, devxdashboard.WeekLabel)
	}
	if periodType == devxdashboard.PeriodQuarterly {
		pr.MonthlyByModel = modelPointsForSubPeriods(read.Events, subject, start, end,
			func(t time.Time) time.Time { return t.AddDate(0, 1, 0) }, devxdashboard.MonthLabel)
	}

	dash, err := devxdashboard.ExportPeriod(pr)
	if err != nil {
		return fmt.Errorf("exporting period dashboard: %w", err)
	}

	output, err := json.MarshalIndent(dash, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling dashboard: %w", err)
	}

	outPath := devxDashboardOutput
	if outPath == "" {
		outPath, err = defaultPeriodReportPath(string(periodType), label)
		if err != nil {
			return fmt.Errorf("resolving default report path: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		return fmt.Errorf("creating reports directory: %w", err)
	}
	if err := os.WriteFile(outPath, output, 0o600); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s report to %s\n", periodType, outPath)

	fmt.Fprintf(os.Stderr, "\nDevX %s report summary:\n", periodType)
	fmt.Fprintf(os.Stderr, "  Person:   %s\n", devxDashboardPerson)
	fmt.Fprintf(os.Stderr, "  Label:    %s\n", label)
	fmt.Fprintf(os.Stderr, "  Period:   %s .. %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	fmt.Fprintf(os.Stderr, "  Events:   %d\n", len(read.Events))
	fmt.Fprintf(os.Stderr, "  Sources:  %d\n", len(r.Sources))
	fmt.Fprintf(os.Stderr, "  Coverage: %.0f%%\n", r.Quality.CoverageScore*100)
	if len(r.Quality.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "  Warnings: %d\n", len(r.Quality.Warnings))
	}

	return nil
}

// periodBoundaries returns the half-open [start, end) UTC range and short
// label (e.g. "2026-W30", "2026-07", "2026-Q3") for the calendar period of
// the given type containing anchor.
func periodBoundaries(periodType devxdashboard.PeriodType, anchor time.Time) (start, end time.Time, label string, err error) {
	anchor = anchor.UTC()
	switch periodType {
	case devxdashboard.PeriodWeekly:
		start = isoWeekStart(anchor)
		end = start.AddDate(0, 0, 7)
		year, week := start.ISOWeek()
		label = fmt.Sprintf("%d-W%02d", year, week)
	case devxdashboard.PeriodMonthly:
		start = time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
		label = start.Format("2006-01")
	case devxdashboard.PeriodQuarterly:
		q := (int(anchor.Month())-1)/3 + 1
		start = time.Date(anchor.Year(), time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 3, 0)
		label = fmt.Sprintf("%d-Q%d", anchor.Year(), q)
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("unknown --period %q (want weekly, monthly, or quarterly)", periodType)
	}
	return start, end, label, nil
}

// isoWeekStart returns the UTC midnight of the Monday starting t's ISO week.
func isoWeekStart(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(t.Weekday()) + 6) % 7 // Monday=0 ... Sunday=6
	return t.AddDate(0, 0, -offset)
}

// modelPointsForSubPeriods buckets [rangeStart, rangeEnd) into successive
// sub-periods (weeks or months, per step) and reduces each with
// report.Build, keeping every observed metric per model so the result can
// back both token and cost stacked-bar charts. Sub-periods with no
// per-model activity are skipped.
func modelPointsForSubPeriods(
	events []omnidevx.Event,
	subject report.Subject,
	rangeStart, rangeEnd time.Time,
	step func(time.Time) time.Time,
	label func(time.Time) string,
) []devxdashboard.ModelPeriodPoint {
	var points []devxdashboard.ModelPeriodPoint
	for cur := rangeStart; cur.Before(rangeEnd); cur = step(cur) {
		subEnd := step(cur)
		if subEnd.After(rangeEnd) {
			subEnd = rangeEnd
		}
		sub := report.Build(events, subject, omnidevx.Period{Start: cur, End: subEnd})
		if len(sub.Metrics.ByModel) == 0 {
			continue
		}
		models := make(map[string]map[string]float64, len(sub.Metrics.ByModel))
		for model, metrics := range sub.Metrics.ByModel {
			vals := make(map[string]float64, len(metrics))
			for k, m := range metrics {
				vals[k] = m.Value
			}
			models[model] = vals
		}
		points = append(points, devxdashboard.ModelPeriodPoint{Label: label(cur), Models: models})
	}
	return points
}

// defaultPeriodReportPath returns the standard on-disk location for a
// period report, matching the path the visionstudio daemon's devxReportsDir
// (cmd/daemon/devx.go) reads from: ~/.plexusone/omnidevx/reports/{type}/{label}.json.
func defaultPeriodReportPath(periodType, label string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".plexusone", "omnidevx", "reports", periodType, label+".json"), nil
}
