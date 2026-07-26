package devxdashboard

import (
	"testing"
	"time"

	omnidevx "github.com/plexusone/omnidevx-core"
	"github.com/plexusone/omnidevx-core/report"
)

func TestExportPeriod_Monthly(t *testing.T) {
	r := &report.DeveloperPeriodReport{
		SchemaVersion: "omnidevx.developer-period/v1",
		Subject:       report.Subject{PersonID: "person:john"},
		Period: omnidevx.Period{
			Start: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
		Metrics: report.MetricSet{
			Combined: map[string]report.Metric{
				"sessions":    {Value: 10},
				"prompts":     {Value: 500},
				"commits":     {Value: 50},
				"input_tokens": {Value: 1_000_000},
				"cost_usd":    {Value: 25.50},
			},
			ByModel: map[string]map[string]report.Metric{
				"claude-opus-4-6": {
					"input_tokens":  {Value: 600_000},
					"output_tokens": {Value: 60_000},
					"cost_usd":      {Value: 15.00},
					"messages":      {Value: 300},
				},
				"claude-haiku-4-5": {
					"input_tokens":  {Value: 400_000},
					"output_tokens": {Value: 40_000},
					"cost_usd":      {Value: 10.50},
					"messages":      {Value: 200},
				},
			},
		},
		Quality: report.DataQuality{CoverageScore: 0.8},
	}

	pr := &PeriodReport{
		Type:   PeriodMonthly,
		Label:  "2026-07",
		Report: r,
		Daily: []DailyPoint{
			{Date: "2026-07-01", Commits: 5, Prompts: 50, CostUSD: 2.50},
			{Date: "2026-07-02", Commits: 3, Prompts: 30, CostUSD: 1.50},
		},
		WeeklyByModel: []ModelPeriodPoint{
			{Label: "W27", Models: map[string]float64{"claude-opus-4-6": 300_000, "claude-haiku-4-5": 200_000}},
			{Label: "W28", Models: map[string]float64{"claude-opus-4-6": 300_000, "claude-haiku-4-5": 200_000}},
		},
	}

	dashboard, err := ExportPeriod(pr)
	if err != nil {
		t.Fatalf("ExportPeriod failed: %v", err)
	}

	if dashboard.ID != "omnidevx-monthly-report" {
		t.Errorf("ID = %q, want omnidevx-monthly-report", dashboard.ID)
	}

	// Should have model data sources
	hasTokensSource := false
	hasCostSource := false
	hasWeeklyTokens := false
	for _, ds := range dashboard.DataSources {
		switch ds.ID {
		case "model-tokens-total":
			hasTokensSource = true
		case "model-cost-total":
			hasCostSource = true
		case "weekly-tokens-by-model":
			hasWeeklyTokens = true
		}
	}

	if !hasTokensSource {
		t.Error("missing model-tokens-total data source")
	}
	if !hasCostSource {
		t.Error("missing model-cost-total data source")
	}
	if !hasWeeklyTokens {
		t.Error("missing weekly-tokens-by-model data source")
	}

	// Should have donut and stacked bar widgets
	hasDonut := false
	hasStackedBar := false
	for _, w := range dashboard.Widgets {
		if w.ID == "tokens-by-model" {
			hasDonut = true
		}
		if w.ID == "weekly-tokens" {
			hasStackedBar = true
		}
	}

	if !hasDonut {
		t.Error("missing tokens-by-model donut widget")
	}
	if !hasStackedBar {
		t.Error("missing weekly-tokens stacked bar widget")
	}
}

func TestExportPeriod_Weekly(t *testing.T) {
	r := &report.DeveloperPeriodReport{
		SchemaVersion: "omnidevx.developer-period/v1",
		Subject:       report.Subject{PersonID: "person:john"},
		Period: omnidevx.Period{
			Start: time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
		},
		Metrics: report.MetricSet{
			Combined: map[string]report.Metric{
				"sessions": {Value: 5},
			},
		},
		Quality: report.DataQuality{CoverageScore: 1.0},
	}

	pr := &PeriodReport{
		Type:   PeriodWeekly,
		Label:  "2026-W28",
		Report: r,
	}

	dashboard, err := ExportPeriod(pr)
	if err != nil {
		t.Fatalf("ExportPeriod failed: %v", err)
	}

	// Weekly should NOT have weekly stacked bar (no sub-period breakdown)
	for _, w := range dashboard.Widgets {
		if w.ID == "weekly-tokens" {
			t.Error("weekly report should not have weekly-tokens widget")
		}
	}
}

func TestWeekLabel(t *testing.T) {
	date := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	got := WeekLabel(date)
	if got != "W28" {
		t.Errorf("WeekLabel = %q, want W28", got)
	}
}

func TestMonthLabel(t *testing.T) {
	date := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	got := MonthLabel(date)
	if got != "Jul" {
		t.Errorf("MonthLabel = %q, want Jul", got)
	}
}
