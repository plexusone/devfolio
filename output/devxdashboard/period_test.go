package devxdashboard

import (
	"encoding/json"
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
			{Label: "W27", Models: map[string]map[string]float64{
				"claude-opus-4-6":  {"input_tokens": 300_000, "cost_usd": 7.50},
				"claude-haiku-4-5": {"input_tokens": 200_000, "cost_usd": 5.00},
			}},
			{Label: "W28", Models: map[string]map[string]float64{
				"claude-opus-4-6":  {"input_tokens": 300_000, "cost_usd": 7.50},
				"claude-haiku-4-5": {"input_tokens": 200_000, "cost_usd": 5.00},
			}},
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
	var weeklyTokens, weeklyCost []map[string]any
	for _, ds := range dashboard.DataSources {
		switch ds.ID {
		case "model-tokens-total":
			hasTokensSource = true
		case "model-cost-total":
			hasCostSource = true
		case "weekly-tokens-by-model":
			if err := json.Unmarshal(ds.Data, &weeklyTokens); err != nil {
				t.Fatalf("unmarshaling weekly-tokens-by-model: %v", err)
			}
		case "weekly-cost-by-model":
			if err := json.Unmarshal(ds.Data, &weeklyCost); err != nil {
				t.Fatalf("unmarshaling weekly-cost-by-model: %v", err)
			}
		}
	}

	if !hasTokensSource {
		t.Error("missing model-tokens-total data source")
	}
	if !hasCostSource {
		t.Error("missing model-cost-total data source")
	}
	if weeklyTokens == nil {
		t.Fatal("missing weekly-tokens-by-model data source")
	}
	if weeklyCost == nil {
		t.Fatal("missing weekly-cost-by-model data source")
	}
	// Tokens and cost must be pulled from their own metrics, not the same
	// underlying values (regression: buildStackedBarData used to ignore
	// its metrics argument and return the same numbers for both charts).
	gotTokens, _ := weeklyTokens[0]["claude-opus-4-6"].(float64)
	gotCost, _ := weeklyCost[0]["claude-opus-4-6"].(float64)
	if gotTokens != 300_000 {
		t.Errorf("weekly tokens for claude-opus-4-6 = %v, want 300000", gotTokens)
	}
	if gotCost != 7.50 {
		t.Errorf("weekly cost for claude-opus-4-6 = %v, want 7.50", gotCost)
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
	// Callers always pass the Monday starting the week (see
	// cmd/devfolio/devx_period.go's isoWeekStart-aligned iteration).
	monday := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	got := WeekLabel(monday)
	if got != "Mon Jul 6, 2026" {
		t.Errorf("WeekLabel = %q, want %q", got, "Mon Jul 6, 2026")
	}
}

func TestMonthLabel(t *testing.T) {
	date := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	got := MonthLabel(date)
	if got != "Jul" {
		t.Errorf("MonthLabel = %q, want Jul", got)
	}
}
